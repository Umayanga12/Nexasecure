import os
import smtplib
from dotenv import load_dotenv
import pyotp
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText


user_otp_counter = {}

def generate_otp(username):
    totp = pyotp.TOTP(pyotp.random_base32(), interval=30)
    otp = totp.now()
    user_otp_counter[username] = {'otp': otp, 'counter': 0}
    return otp
