import flask

from flask import Flask, jsonify, request
from blockchain import Blockchain
from hardwarewallet import HardwareWallet
from DOT import DOT_Pool
from pymongo import MongoClient
from change_own import AuthenticationSystem


Blockchain = Blockchain()
Wallet = HardwareWallet()
DOT_Pool = DOT_Pool()
AuthenticationSystem = AuthenticationSystem(DOT_Pool)

app = Flask(__name__)

mongoClient = MongoClient('mongodb://localhost:27017/')
database = mongoClient['ownership-token-db']


@app.route('/login', methods=['POST'])
def login():
    data = flask.request.get_json()
    required = ['username', 'uuid']
    if not all(k in data for k in required):
        return 'Missing values', 400
    
    index = Blockchain.new_auth_request(data['username'],data['uuid'])
    authenticated = AuthenticationSystem.authenticate(data['username'],data['username'])
    if authenticated:
        Wallet.transfer_DOT(index,"company")
        sinded_auth = Wallet.sign_auth_data(index)
        return sinded_auth,200
    return False

@app.route('/logout', methods=['POST'])
def logout():
    metadata = {
        'user':'user',
    }

    DOT = DOT_Pool.mint_DOT(metadata)
    Wallet.store_DOT(DOT)
    return True

#get all the data from blockchain
@app.route("/chain",methods=['GET'])
def full_chain():
    response = {
        'chain': Blockchain.chain,
        'length': len(Blockchain.chain)
    }
    return flask.jsonify(response), 200


@app.route('/passtoken/<dip_id>', methods=['PUT'])
def passtoken():
    # Find a token with status 'not_inuse'
    token_to_update = collection.find_one({"status": "not_inuse"})
    
    if not token_to_update:
        return jsonify({'message': 'No token with status "not_inuse" found'}), 404
    
    # Update the status to 'inuse'
    updated_token = collection.update_one(
        {"_id": token_to_update["_id"]},
        {"$set": {"status": "inuse"}}
    )
    
    if updated_token.modified_count > 0:
        return jsonify({
            'message': 'Token status updated successfully',
            'token_id': str(token_to_update["_id"]),
            'new_status': 'inuse'
        }), 200
    else:
        return jsonify({'message': 'Failed to update token status'}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
