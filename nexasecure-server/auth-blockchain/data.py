
import pymongo
from pymongo import MongoClient

urls = ["mongodb://localhost:27017/","mongodb://localhost:27018/"]
dbs = ["ownership-token-db","request_token_pool"]

def add_data():
    for url in urls:
        for db in dbs:
            dbclient = MongoClient(url)
            db = dbclient[db]
            department = ["dep_2", "dep_3", "dep_4", "dep_5"]

            for dep in department:
                collection = db[dep]
                tokens = [
                    {"token_id": 1, "owner_address": "{dep}_0xabc123", "status": "NOT_INUSE"},
                    {"token_id": 2, "owner_address": "{dep}_0xdef456", "status": "NOT_INUSE"},
                    {"token_id": 3, "owner_address": "{dep}_0xghi789", "status": "NOT_INUSE"}
                ]
                result = collection.insert_many(tokens)
                print("Inserted IDs:", result.inserted_ids)



