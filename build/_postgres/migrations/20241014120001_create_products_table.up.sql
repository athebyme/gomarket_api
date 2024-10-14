CREATE TABLE IF NOT EXISTS products (
    global_id SERIAL PRIMARY KEY,
    appellation VARCHAR(255) NOT NULL,
    categoryID INT NOT NULL,
    distance FLOAT NOT NULL
);
