# use mhkit to gather some wave data
from mhkit.wave.io import ndbc
from mhkit.wave import resource

# database connection
import sqlalchemy

# io operations
import io

# dataframes
import pandas as pd


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


param = "stdmet"
# gather available data - make a random selection of buoys
available_buoys = ndbc.available_data(parameter=param)

# dataframe has columns of "id", "year", "filename" describing available data in the swden data set form ndbc
available_buoy_ids = available_buoys["id"].unique()
# define our desired minimum span of years
year_span = 10

# array of dataframe sections that meet year span requirement
found = []
for buoy_id in available_buoy_ids:
    # this buoy has bad formatted data throwing errors from MHKiT when parsing the gzip text file requested
    if buoy_id == "41001":
        continue
    d = available_buoys.loc[available_buoys["id"] == buoy_id]
    years = d["year"].unique()
    if len(years) >= year_span:
        found.append(d)

# want to collect 0.5 gb (1000mb) of data
mem_used = 0
mem_desired = 500
# dict of data sets collected
data_sets = {}
# renaming for sql inserts & requried columns
REQUIRED_COLUMNS = set(["WVHT", "DPD"])
RENAME = {"WVHT": "hm0", "DPD": "tp"}

# gather data
for index, df in enumerate(found):
    if mem_used >= mem_desired:
        break
    print(df["id"].unique())
    print(mem_used)
    filenames = df["filename"]
    # get dictionary of dataframes per year of available data -> dict keys are the years
    try:
        data_dict = ndbc.request_data(parameter=param, filenames=filenames)
    # handle errors but allow keyboard interrupt to interupt the loop
    # some data sets aren't formatted correctly and throw errors on parsing the compressed zip file
    except KeyboardInterrupt:
        break
    except:
        continue

    dfs = []

    for year in data_dict:
        data_dict[year] = ndbc.to_datetime_index(parameter=param, ndbc_data=data_dict[year])
        dfs.append(data_dict[year])
    try:
        # some data sources may not return proper columns
        data = pd.concat(dfs)
    except KeyboardInterrupt:
        break
    except:
        continue

    try:
        # remove unit row
        data = data[data["WVHT"] != "m"]
        # only columns we want
        data = data[list(REQUIRED_COLUMNS)]
        # float
        data[list(REQUIRED_COLUMNS)] = data[list(REQUIRED_COLUMNS)].astype(float)
        # sort df
        data = data.sort_index()
        # rename columns
        data = data.rename(columns=RENAME)
    except KeyboardInterrupt:
        break
    except:
        continue

    data[["hm0", "tp"]] = data[["hm0", "tp"]].astype(float)  # make sure data is numeric
    data["date"] = data.index
    data = data.reset_index(drop=True)
    data["data_set_id"] = index + 1
    data["date"] = pd.to_datetime(data["date"], utc=True)
    data_sets[index + 1] = {}
    data_sets[index + 1]["data"] = data
    data_sets[index + 1]["data_size"] = data.memory_usage(deep=True).sum() / 1e6
    mem_used += data.memory_usage(deep=True).sum() / 1e6


# binary insert query - sqlalchemy bindvars
q_insert = """
    INSERT into binary_data (data_set_id, data)
    VALUES
    (:data_set_id, :data)
"""
# size insert
q_size = """
    INSERT INTO data_size (data_set_id, data_size_mb)
    VALUES
    (:data_set_id, :data_size_mb)
"""

# No args as EGI setup works with default options
pool = connect_pg()


# context manager to close the pool connection
with pool.connect() as c:
    for key in data_sets:
        # display info on each insert
        print(data_sets[key]["data"].head())
        print("Above table using {} MB".format(data_sets[key]["data_size"]))
        data = data_sets[key]["data"]
        # insert the row data
        data.to_sql(name="row_data", con=c, if_exists="append", index=False)
        # remove data_set_id from df - only storing the csv data in bytes
        data = data.drop(columns=["data_set_id"])
        # prepare bytes
        buf = io.StringIO()
        data.to_csv(buf)
        # reset buffer to start point
        buf.seek(0)
        rs = c.execute(
            sqlalchemy.text(q_insert),
            {
                "data_set_id": key,
                "data": buf.read(),  # reason we needed to seek to 0 point
            },
        )
        # insert size - easier for post analysis to re-grab the size
        rs = c.execute(
            sqlalchemy.text(q_size),
            {"data_set_id": key, "data_size_mb": data_sets[key]["data_size"]},
        )


# now run some queries to look at the data!
