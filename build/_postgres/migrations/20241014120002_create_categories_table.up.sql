CREATE TABLE IF NOT EXISTS categories (
    categoryID SERIAL PRIMARY KEY,
    category VARCHAR(255) NOT NULL,
    parentCategoryID INT,
    parentCategoryName VARCHAR(255)
);