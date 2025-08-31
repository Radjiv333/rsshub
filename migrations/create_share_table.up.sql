CREATE TABLE share (
    id SERIAL PRIMARY KEY CHECK (id = 1),
    interval TEXT NOT NULL,
    workers_num INT NOT NULL
);