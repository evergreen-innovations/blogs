# db connection
import sqlalchemy

# io operations
import io

# dataframes
import pandas as pd

# time operations (simple method)
import time

# random operations
import random

# EGI local works with default connection - modify entries as needed
pool_config = {}
# pool_config would look like if you have custom options you'd like to test
# {
#     "pool_size": 1,  # max num permanent connections
#     "max_overflow": 1,  # temp exceed pool size if no connections available
#     "pool_timeout": 30,  # max seconds to wait until when retrieving new connection from pool
#     "pool_recycle": 3600,  # max seconds a pool connection can persist
# }
# EGI local set up works with default connection
db_params = {}
# connection would look like if you have custom options
# {
#     "DB_USER": "postgres",
#     "DB_PASS": "postgres",
#     "DB_HOSTNAME": "localhost",
#     "DB_PORT": 5432,
#     "DB_NAME": "speed-blog",
# }


def connect_pg(db_params: dict = None, pool_config: dict = None) -> sqlalchemy.engine.Engine:
    """connects to a postgres database instance

    Args:
        db_params (dict): Dictionary with following keys - any can be left out:
            - DB_USER: str - user to login to db with
            - DB_PASS: str - password for user
            - DB_HOSTNAME: str - localhost or remote host url
            - DB_PORT: int - 5432
            - DB_NAME: str - name of database to connect to
        pool_config (dict): Dictionary with following keys - any can be left out:
            - pool_size: int - max num permanent connections
            - max_overflow: int - temp exceed pool size if no connections available
            - pool_timeout: int - max seconds to wait until when retrieving new connection from pool
            - pool_recycle: int - max seconds a pool connection can persist


    Returns:
        sqlalchemy.engine.Engine: connection pool used in various querying methods
    """
    default_connection = {
        "DB_USER": "postgres",
        "DB_PASS": "postgres",
        "DB_HOSTNAME": "localhost",
        "DB_PORT": 5432,
        "DB_NAME": "speed-blog",
        "drivername": "postgresql",
    }
    params = {}
    if db_params is not None:
        for key in default_connection:
            # use default connection options if missing in db_params arg
            v = db_params.get(key, default_connection[key])
            params[key] = v
    else:
        params = default_connection

    default_pool = {
        "pool_size": 1,
        "max_overflow": 1,
        "pool_timeout": 30,
        "pool_recycle": 3600,
    }

    pool_params = {}
    if pool_config is not None:
        for key in default_pool:
            # use default connection options if missing in pool_config arg
            v = pool_config.get(key, default_pool[key])
            pool_params[key] = v
    else:
        pool_params = default_pool
    # create the connection pool
    pool = sqlalchemy.create_engine(
        sqlalchemy.engine.url.URL(
            drivername=params["drivername"],
            username=params["DB_USER"],
            password=params["DB_PASS"],
            host=params["DB_HOSTNAME"],
            port=params["DB_PORT"],
            database=params["DB_NAME"],
        ),
        **pool_params,
    )

    return pool


data_log_rows = {
    "set_id": [],
    "min_date": [],
    "max_date": [],
    "time_to_request_all_data": [],
    "tp_sum": [],
}

data_log_binary = {
    "set_id": [],
    "min_date": [],
    "max_date": [],
    "time_to_request_all_data": [],
    "tp_sum": [],
}

# 58 data sets
availabe_ids = list(range(1, 58))
row_q = """
SELECT * FROM row_data where data_set_id = {}
"""
binary_q = """
SELECT * from binary_data where data_set_id = (:set_id)
"""
# no options needed for EGI connection parameters
pool = connect_pg()

req_query = """
SELECT data_set_id, min(date), max(date)
from row_data
group by data_set_id;
"""

with pool.connect() as c:
    req_info = pd.read_sql(sql=req_query, con=c)

requests = []

# build request parameters for filtering the row data
for d in req_info.iterrows():
    start = pd.to_datetime(d[1]["min"]).year
    end = pd.to_datetime(d[1]["max"]).year
    # sample two numbers between start and end years for random duration to filter down to in testing
    years = random.sample(range(start, end), 2)
    requests.append({"id": d[1]["data_set_id"], "start_year": min(years), "end_year": max(years)})

# gather all data - no filtering - for analysis
with pool.connect() as c:
    for id in availabe_ids:
        print(f"Getting all data for data_id {id}")
        # row data
        start = time.time()
        d_rows = pd.read_sql(sql=row_q.format(id), con=c)
        data_log_rows["set_id"].append(d_rows["data_set_id"].unique())
        end = time.time() - start
        data_log_rows["time_to_request_all_data"].append(f"{end} seconds")
        d_rows["date"] = pd.to_datetime(d_rows["date"], utc=True)
        min_date = d_rows.date.min()
        max_date = d_rows.date.max()
        tp_sum = d_rows["tp"].sum()
        data_log_rows["tp_sum"].append(tp_sum)
        data_log_rows["max_date"].append(max_date)
        data_log_rows["min_date"].append(min_date)

        # binary data
        start = time.time()
        rs = c.execute(sqlalchemy.text(binary_q), {"set_id": id})
        res = rs.fetchall()
        d = io.BytesIO(res[0]["data"])  # binary data
        d_binary = pd.read_csv(d)
        end = time.time() - start
        data_log_binary["time_to_request_all_data"].append(f"{end} seconds")
        d_binary["date"] = pd.to_datetime(d_binary["date"], utc=True)
        min_date = d_binary.date.min()
        max_date = d_binary.date.max()
        data_log_binary["max_date"].append(max_date)
        data_log_binary["min_date"].append(min_date)
        data_log_binary["tp_sum"].append(d_binary["tp"].sum())

        d_id = res[0]["data_set_id"]  # one row per data set so okay to index this way
        data_log_binary["set_id"].append(d_id)


data_log_binary = pd.DataFrame(data_log_binary)
data_log_rows = pd.DataFrame(data_log_rows)

data_log_rows.to_csv("all_years_rows.csv", index=False)
data_log_binary.to_csv("all_years_binary.csv", index=False)

# new query for fow data - binary still the same
row_q = """
SELECT * from row_data WHERE data_set_id = {id}
and extract(YEAR from date) >= {start_year}
and extract(YEAR from date) <= {end_year}
"""

# do same as above but with filtering in the time frame - request one year at a time
data_log_rows = {
    "set_id": [],
    "min_date": [],
    "max_date": [],
    "time_to_gather_data": [],
    "tp_sum": [],
}

data_log_binary = {
    "set_id": [],
    "min_date": [],
    "max_date": [],
    "time_to_gather_data": [],
    "tp_sum": [],
}

with pool.connect() as c:
    for param in requests:
        print(f"Getting data with params {param}")
        # row data
        start = time.time()
        # can unpack the param dictionary based off the format var names in query
        d_rows = pd.read_sql(sql=row_q.format(**param), con=c)
        data_log_rows["set_id"].append(d_rows["data_set_id"].unique())
        end = time.time() - start
        data_log_rows["time_to_gather_data"].append(f"{end} seconds")
        d_rows["date"] = pd.to_datetime(d_rows["date"], utc=True)
        min_date = d_rows.date.min()
        max_date = d_rows.date.max()
        tp_sum = d_rows["tp"].sum()
        data_log_rows["tp_sum"].append(tp_sum)
        data_log_rows["max_date"].append(max_date)
        data_log_rows["min_date"].append(min_date)

        # binary data
        start = time.time()
        rs = c.execute(sqlalchemy.text(binary_q), {"set_id": param["id"]})
        res = rs.fetchall()
        d = io.BytesIO(res[0]["data"])  # binary data
        d_binary = pd.read_csv(d)
        d_binary["date"] = pd.to_datetime(d_binary["date"], utc=True)
        # need to filter dataframe in memory here
        mask = (d_binary["date"].dt.year >= param["start_year"]) & (
            d_binary["date"].dt.year <= param["end_year"]
        )
        d_binary = d_binary[mask]
        end = time.time() - start
        data_log_binary["time_to_gather_data"].append(f"{end} seconds")

        min_date = d_binary.date.min()
        max_date = d_binary.date.max()
        data_log_binary["max_date"].append(max_date)
        data_log_binary["min_date"].append(min_date)
        data_log_binary["tp_sum"].append(d_binary["tp"].sum())

        d_id = res[0]["data_set_id"]  # one row per data set so okay to index this way
        data_log_binary["set_id"].append(d_id)


data_log_binary = pd.DataFrame(data_log_binary)
data_log_rows = pd.DataFrame(data_log_rows)

data_log_rows.to_csv("filtered_years_rows.csv", index=False)
data_log_binary.to_csv("filtered_years_binary.csv", index=False)
