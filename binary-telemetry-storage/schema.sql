-- SQL file to generate the tables needed for this experiment
-- NOTE - if you re-run the insert python code - re-run this file to reset tables
DROP TABLE IF EXISTS row_data;
CREATE TABLE row_data (
    data_set_id int,
    date timestamptz not null,
    Hm0 double precision,
    Te double precision,
    Tp double precision
);


DROP TABLE IF EXISTS binary_data;
CREATE TABLE binary_data(
    data_set_id int,
    data bytea -- bytea is data type for binary in PostgreSQL
);


DROP TABLE IF EXISTS data_size;
CREATE TABLE data_size (
    data_set_id int,
    data_size_mb double precision
);

SET TIME ZONE 'UTC'; -- makes our dates align between csv binary and datetime objects in db for row data

-- Example queries to look at data

-- SELECT * FROM row_data WHERE data_set_id = 1;
-- SELECT * FROM binary_data;
-- SELECT * FROM data_size;