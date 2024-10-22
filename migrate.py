import psycopg2
from pymongo import MongoClient

# Подключение к MongoDB
mongo_client = MongoClient("mongodb+srv://athebyme:fbPifkWZs4u8BPKQ@cluster0.s3lshse.mongodb.net/wholesaler-products")
mongo_db = mongo_client["wholesaler-products"]
mongo_products_collection = mongo_db["all-products"]
mongo_descriptions_collection = mongo_db["all-products-descriptions"]
mongo_prices_stock_collection = mongo_db["all-products-prices-stocks"]
mongo_wb_collection = mongo_client['wildberries']['categories']

# Подключение к PostgreSQL (в Docker)
pg_conn = psycopg2.connect(
    dbname="postgres",
    user="postgres",
    password="postgres",
    host="localhost",
    port="5432"
)
pg_cursor = pg_conn.cursor()

global_ids = []
categories = {}
# Вставка данных из MongoDB в PostgreSQL
def insert_data():
    # Перенос данных из коллекции продуктов
    for product in mongo_products_collection.find():
        global_id = product.get("global_id")
        model = product.get("model", "")
        appellation = product.get("appellation", "")
        category = product.get("category", "")
        brand = product.get("brand", "")
        country = product.get("country", "")
        product_type = product.get("product_type", "")
        features = product.get("features", "")
        sex = product.get("sex", "")
        color = product.get("color", "")
        dimensions = product.get("dimensions", "")
        package = product.get("package", "")
        media = product.get("photo_codes", "")
        barcodes = product.get("barcodes", "")
        material = product.get("material", "")
        package_battery = product.get("package_battery", "")
        global_ids.append(global_id)

        # SQL-запрос для вставки данных
        pg_cursor.execute("""
            INSERT INTO wholesaler.products (global_id, model, appellation, category, brand, country, product_type, features, sex, color, dimension, package, media, barcodes, meterial, package_battery)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (global_ID) DO NOTHING
        """, (global_id, model, appellation, category, brand, country, product_type, features, sex, color, dimensions,
              package, media, barcodes, material, package_battery))
    pg_conn.commit()
    # Перенос данных из коллекции описаний
    for description in mongo_descriptions_collection.find():
        global_id = description.get("global_id")
        product_description = description.get("product_description", "")
        product_appellation = description.get("product_appellation", "")

        if global_id not in global_ids:
            print(f'{global_id} not in global_ids for products!')
            continue
        # SQL-запрос для вставки описаний
        pg_cursor.execute("""
            INSERT INTO wholesaler.descriptions (global_id, product_description, product_appellation)
            VALUES (%s, %s, %s)
        """, (global_id, product_description, product_appellation))
    pg_conn.commit()

    # Перенос данных из коллекции цен и стоков
    for price_stock in mongo_prices_stock_collection.find():
        global_id = price_stock.get("global_id")
        main_articular = price_stock.get("main_articular", "")
        price = price_stock.get("price", 0)
        stock = 1 if price_stock.get("stock") == "+" else 0
        if global_id not in global_ids:
            print(f'{global_id} not in global_ids for products!')
            continue

        # Вставка стоков
        pg_cursor.execute("""
            INSERT INTO wholesaler.stocks (global_id, main_articular, stocks)
            VALUES (%s, %s, %s)
        """, (global_id, main_articular, stock))

        # Вставка цен
        pg_cursor.execute("""
            INSERT INTO wholesaler.price (global_id, price)
            VALUES (%s, %s)
        """, (global_id, price))
    pg_conn.commit()
    for wildberries in mongo_wb_collection.find():
        global_id = wildberries.get("global_id")
        appellation = wildberries.get("appellation", "")
        category = wildberries.get("category", {})
        distance = wildberries.get("distance", 0.0)
        category_id = category.get("categoryID")
        category_name = category.get("category")
        parent_category_id = category.get("parentCategoryID")
        parent_category_name = category.get("parentCategoryName")
        if global_id not in global_ids:
            print(f'{global_id} not in global_ids for products!')
            continue

        if category_id not in categories.keys():
            categories[category_id] = {'category' : category_name,
                                       'parent_category_id' : parent_category_id,
                                       'parent_category_name' : parent_category_name}
        pg_cursor.execute("""
               INSERT INTO wildberries.products (global_id, appellation, category_id, distance)
               VALUES (%s, %s, %s, %s)
               ON CONFLICT (global_id) DO NOTHING
           """, (global_id, appellation, category_id, distance))
    for k,v in categories.items():
        pg_cursor.execute("""
               INSERT INTO wildberries.categories (category_id, category, parent_category_id, parent_category_name)
               VALUES (%s, %s, %s, %s)
           """, (k, v['category'], v['parent_category_id'], v['parent_category_name']))
    pg_conn.commit()




# Выполнить вставку данных
insert_data()

# Закрыть соединения
pg_cursor.close()
pg_conn.close()
mongo_client.close()
