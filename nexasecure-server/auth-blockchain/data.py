#adding mok  data to the database 


###for the testing purpose only

from pymongo import MongoClient

dbclient = MongoClient('mongodb://localhost:27017/')
db = dbclient['ownership-token-db']
department = ["dep_2", "dep_3", "dep_4", "dep_5"]

for dep in department:
    collection = db[dep]
    tokens = [
        {"token_id": 1, "owner_address": "0xabc123", "status": "NOT_INUSE"},
        {"token_id": 2, "owner_address": "0xdef456", "status": "NOT_INUSE"},
        {"token_id": 3, "owner_address": "0xghi789", "status": "NOT_INUSE"}
    ]
    result = collection.insert_many(tokens)
    print("Inserted IDs:", result.inserted_ids)