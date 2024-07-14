from genarate_OTP import generate_otp, user_otp_counter
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/gen-otp', methods=['POST'])
def gen_otp():
    username = request.json.get("username")

    if not username:
        return jsonify({'message':'username is required'}),400
    
    
    otp = generate_otp(username)
    user_otp_counter[username] = {'otp': otp,'counter': 0}
    print(otp)
    return jsonify({'message': 'OTP generated successfully'}), 200



@app.route('/verify-otp', methods=['POST'])
def validate_otp():
    username = request.json.get("username")
    otp_attemp = request.json.get("otp")

    if username not in user_otp_counter:
        return jsonify({'Error':'OTP is not generated for the user'}),400
    
    user_data = user_otp_counter[username]
    
    if user_data['counter'] >=3:
        return jsonify({'Error':'Maximum attempts reached'}),403
    
    if user_data['otp'] == otp_attemp:
        return jsonify({'message':'OTP verified successfully'}),200
    else:
        user_data['counter'] += 1
        return jsonify({'Error':'Invalid OTP'}),400
    


if __name__ == '__main__':
    app.run(debug=True,port=5000)