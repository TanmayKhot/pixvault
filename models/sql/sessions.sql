CREATE TABLE SESSIONS (
    id serial PRIMARY KEY,
    user_id INT UNIQUE,
    toekn_hash  TEST UNIQUE NOT NULL 
)