# Maps in Go

## Introduction
C++ was the first general purpose programming language I learnt
(after MATLAB) and where I was introduced to the concept of [hash
tables](https://en.wikipedia.org/wiki/Hash_table). Hash tables have
many different names but here we will refer to them as a `map`. `Maps`
are a fundamental data structure as they
provide a simple interface (and constant-time lookup) for retrieving
specific elements of stored data. Each element in the `map` is a
key-value pair. `Maps` may also be iterated over to
view each key-value pair.

After having segfaults in some C++ code, I learnt the hard way about
"iterator invalidation" when an element is deleted from the map
during iteration.

Later, coming to Go two things struck me immediately about the
in-built `map` type that were different compared to C++:
1) Deleting during iteration didn't require any special thought
2) The iteration order is non-deterministic.

This blog explores the implementations in both languages to explain
these differences. When source code is presented, comments added for
this blog are prefaced `[Blog]`.


## C++ implementation
C++ provides
[the](https://en.cppreference.com/w/cpp/container/unordered_map)
`unordered_map` type for storing key-value pairs. There are a number
of different implementations of the C++ Standard Libary, here we will be
looking at the `libc++` [implementation](https://libcxx.llvm.org/).

The `libc++` `unordered_map`
[implementation](https://github.com/llvm/llvm-project/blob/main/libcxx/include/unordered_map)
uses the `hash` of the key is to index into an
array, known as a bucket, and each bucket is a linked list which
stores the individual key-value pairs.

Internally the `unordered_map` defers to the `__hash_table`
[class](https://github.com/llvm/llvm-project/blob/main/libcxx/include/__hash_table),
where we can see this structure in the `__bucket_list`
[type](https://github.com/llvm/llvm-project/blob/62c481542e63a9019aa469c70cb228fe90ce7ece/libcxx/include/__hash_table#L949):

```c++
typedef unique_ptr<__next_pointer[], __bucket_list_deleter> __bucket_list;
```

The `__bucket_list` is therefore an array of pointers, each of these
points is to the first element in the linked list that holds the
key-value pairs, as shown below.

![C++ unordered_map data structure](images/c++.png "C++ unordered_map
data structure")

There is some interesting discussion in the
[proposal](http://www.open-std.org/jtc1/sc22/wg21/docs/papers/2003/n1456.html)
for introducing the type into the C++ Standard Library on the design choices.

We can see this structure in the `find`
[method](https://github.com/llvm/llvm-project/blob/62c481542e63a9019aa469c70cb228fe90ce7ece/libcxx/include/__hash_table#L2378)
(comments mine):

```c++
template <class _Tp, class _Hash, class _Equal, class _Alloc>
template <class _Key>
typename __hash_table<_Tp, _Hash, _Equal, _Alloc>::iterator
__hash_table<_Tp, _Hash, _Equal, _Alloc>::find(const _Key& __k)
{
    size_t __hash = hash_function()(__k);           // [Blog] Get the hash for the key
    size_type __bc = bucket_count();
    if (__bc != 0)                                  // [Blog] Check we have buckets
    {
        size_t __chash = __constrain_hash(__hash, __bc); // [Blog] Turn the hash into an index
        __next_pointer __nd = __bucket_list_[__chash];   // [Blog] Get the pointer to first element in the linked list
        if (__nd != nullptr)
        {
            // Walk the linked list: follow the __next_ pointer,
            // but only in this bucket (more on this below).
            for (__nd = __nd->__next_; __nd != nullptr &&
                (__nd->__hash() == __hash
                  || __constrain_hash(__nd->__hash(), __bc) == __chash);
                                                           __nd = __nd->__next_)
            {
            // Check if this element matches the one we are looking for
                if ((__nd->__hash() == __hash)
                    && key_eq()(__nd->__upcast()->__value_, __k))
                    return iterator(__nd);         // [Blog] Return the element
            }
        }
    }

    // Element not found
    return end();
}
```

A crucial part of the C++ implementation is that __final__ element in the
linked list points to the __next__ element in the
`__bucket_list`. This explains why in `find` there is a check to make sure we
stay in the same bucket as we follow the `__next` pointer in the
linked list.

This means that for iteration we just keep following
the `__next` pointer until we reach the end of the `_bucket_list` and,
therefore, the iteration order identical (assuming no elements are
removed or added) between iterations.

This underlying data structure also explains why the following program
segfaults. When the current element (`it`) is removed, the chain of
`__next` pointers is broken.

```c++
#include <iostream>
#include <string>
#include <unordered_map>

int main() {
  std::unordered_map<std::string, int> m = {
    {"ABC", 1},
    {"DEF", 2},
    {"GHI", 3}
  };

  for (auto it = m.begin(); it != m.end();) {
    if (it->second %2 == 0) {
      m.erase(it); // Segfault!
    } else {
      ++it;
    }
  }
}
```

The fix is a simple one, as of C++11, the `erase`
[method](https://en.cppreference.com/w/cpp/container/unordered_map/erase)
returns the `__next` pointer of the deleted element to allow the iteration to continue.

```c++
it = m.erase(it)
```


## Go implementation
The implementation in Go follows a similar strategy but has some
fundamental differences. There is still an array with a pointer to a
list of key-value pairs (a bucket) but in Go's case each bucket holds
[eight
elements](https://github.com/golang/go/blob/1e34c00b4c84a32423042e3d03397277e6c3573c/src/runtime/map.go#L67). If
a particular bucket needs to store more than eight elements, initially
an "overflow bucket" is used until a threshold is hit, which triggers
twice as many buckets to be allocated and all the existing elements
are rearranged into the new buckets.

Another
interesting facet is that within the bucket all the keys are stored first,
[followed by the
values](https://github.com/golang/go/blob/1e34c00b4c84a32423042e3d03397277e6c3573c/src/runtime/map.go#L155)
to be maximally space efficient. Vincent Blanchon has a nice overview
of the Go map type in a
[two](https://medium.com/a-journey-with-go/go-map-design-by-example-part-i-3f78a064a352)
[part](https://medium.com/@blanchon.vincent/go-map-design-by-code-part-ii-50d111557c08)
series which covers the underlying data structure, including the
overflow buckets and key-value packing.

In the `mapAccess1`
[function](https://github.com/golang/go/blob/1e34c00b4c84a32423042e3d03397277e6c3573c/src/runtime/map.go#L395)
we can see a similar routine for finding
a given
element. [First](https://github.com/golang/go/blob/1e34c00b4c84a32423042e3d03397277e6c3573c/src/runtime/map.go#L419)
the key is hashed and converted to an index in the bucket array:

```go
	hash := t.hasher(key, uintptr(h.hash0))
	m := bucketMask(h.B)
	b := (*bmap)(add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
```

and then the bucket is
[iterated](https://github.com/golang/go/blob/1e34c00b4c84a32423042e3d03397277e6c3573c/src/runtime/map.go#L432)
to find the corresponding key,

```go
    top := tophash(hash)
bucketloop:
	for ; b != nil; b = b.overflow(t) {            // [Blog] Take care of overflow buckets
		for i := uintptr(0); i < bucketCnt; i++ {  // [Blog] Walk the bucket
			if b.tophash[i] != top {
				if b.tophash[i] == emptyRest {
					break bucketloop
				}
				continue
			}
			k := add(unsafe.Pointer(b),	dataOffset+i*uintptr(t.keysize))    // [Blog] Get the key from the bucket data structure
			if t.indirectkey() {
				k = *((*unsafe.Pointer)(k))
			}
			if t.key.equal(key, k) {           // [Blog] Is this the key?
				e := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
				if t.indirectelem() {
					e = *((*unsafe.Pointer)(e))
				}
				return e                       // [Blog] Return the value
			}
		}
	}
	return unsafe.Pointer(&zeroVal[0])           // [Blog] Key not found
```

Although the data structure is slightly different, so far the C++ and
Go maps have a similar approach for finding the value for a given
key. The iteration behaviour is quite
different and can be a suprise to those new to Go; in the Section "[For
Statemets with range clause](https://go.dev/ref/spec#For_statements)", we have

> 3. The iteration order over maps is not specified and is not
>    guaranteed to be the same from one iteration to the next.

This is a deliberate design choice in order to stop developers relying
on a particular interation order. If the Go Team wanted to change
the underlying implementation of the `map` type (but keep the same
API), a change in the iteration order could then, perhaps silently,
break existing code. Early on, the order was deterministic if the map
had fewer than eight elements (ie 1 bucket) but this too was made
non-deterministic ([Issue
6719](https://github.com/golang/go/issues/6719)).

This iteration order begins in the `mapiterinit`
[function](https://github.com/golang/go/blob/1e34c00b4c84a32423042e3d03397277e6c3573c/src/runtime/map.go#L815)

```go
	// decide where to start
	r := uintptr(fastrand())             // [Blog] Get a random number
	if h.B > 31-bucketCntBits {
		r += uintptr(fastrand()) << 31
	}
	it.startBucket = r & bucketMask(h.B) // [Blog] Get starting index from the random number
	it.offset = uint8(r >> h.B & (bucketCnt - 1)) // [Blog] Start at a random element within the bucket (issue 6719)

	// iterator state
	it.bucket = it.startBucket
```

This means we start the iteration at a random bucket and at a random
point within the bucket. In essence we start with a random
key-pair. Afterwards each key-pair is visited until we reach the end
of the bucket list, at which point we go back to the start of the
bucket list and continue to visit the elements until we are back where
we started (`it.startBucket`).

The logic in the `mapiternext`
[function](https://github.com/golang/go/blob/1e34c00b4c84a32423042e3d03397277e6c3573c/src/runtime/map.go#L864)
is quite involved so some of the code is elided.

```go

next:
	if b == nil {  // [Blog] Reached the end of a bucket (overflow is nil)
		// [Blog] We have arrived back at the start bucket after wrapping around
		if bucket == it.startBucket && it.wrapped {
			// end of iteration
			it.key = nil
			it.elem = nil
			return
		}

		// [Blog] (snip)
		bucket++                         // [Blog] Move the the next bucket
		if bucket == bucketShift(it.B) { // [Blog] Go back to the start
			bucket = 0
			it.wrapped = true            // [Blog] Note wrapped being set to true
		}
		i = 0
	}
	for ; i < bucketCnt; i++ {          // [Blog] Loop through the buckets
		offi := (i + it.offset) & (bucketCnt - 1)  // [Blog] The key-pair to visit involves the random offset (it.offset)
		if isEmpty(b.tophash[offi]) || b.tophash[offi] == evacuatedEmpty {
			// TODO: emptyRest is hard to use here, as we start iterating
			// in the middle of a bucket. It's feasible, just tricky.
			continue
		}

		// [Blog] Get the key
		k := add(unsafe.Pointer(b), dataOffset+uintptr(offi)*uintptr(t.keysize))
		if t.indirectkey() {
			k = *((*unsafe.Pointer)(k))
		}

		// [Blog] Get the value
		e := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+uintptr(offi)*uintptr(t.elemsize))

		// [Blog] (snip)

		if (b.tophash[offi] != evacuatedX && b.tophash[offi] != evacuatedY) ||
			!(t.reflexivekey() || t.key.equal(k, k)) {
			// This is the golden data, we can return it.
			// [Blog] (snip)
			it.key = k                                   // [Blog] Set the key
			if t.indirectelem() {
				e = *((*unsafe.Pointer)(e))
			}
			it.elem = e                                  // [Blog] Set the value
		} else {
			// [Blog] (snip)
		}

		it.bucket = bucket
		if it.bptr != b { // avoid unnecessary write barrier; see issue 14921
			it.bptr = b
		}
		it.i = i + 1
		it.checkBucket = checkBucket
		return            // [Blog] return to yield key and value, all state is stored in the iterator
	}
	b = b.overflow(t) // [Blog] check for overflow buckets
	i = 0
	goto next         // [Blog]  start the iteration procedure
```

With this in mind, it's not true to say that the iteration order is
"random". We proceed through the buckets in a specific order, it is
just that from one iteration to the next, the starting bucket and the
order through each bucket will be different.

This iteration strategy also answers the question we had at the start
of this post about why deletion does not require any special
thought. We know that each of the buckets is always eight
elements large (with potentially an overflow bucket) which means that
we do not have the chain of `next` pointers to keep a track of. The
[Spec](https://go.dev/ref/spec#For_statements) also says

> If a map entry that has not yet been reached is removed during
> iteration, the corresponding iteration value will not be
> produced. If a map entry is created during iteration, that entry may
> be produced during the iteration or may be skipped.

This aligns with what we have seen. If an element is created during
iteration it will appear in a particular bucket depending on the hash
of the key. Whether we see it not during iteration depends on whether
we have already visited that bucket. The same logic applies for
deletion.


## Conclusion
We have looked in detail at the implementations of a `map` data
structure in both C++ and Go. They both have similar functionality
— insert, delete, iterate — but the underlying implementation yields
different behaviour. The non-deterministic iteration order in Go is
particularly interesting and shows the length to which the Go team
will go to provide backwards compatibilty. In their [blog
post](https://go.dev/blog/maps) on maps, an example is given for
providing a deterministic interation order.
