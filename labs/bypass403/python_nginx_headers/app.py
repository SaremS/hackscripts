from flask import Flask, request, abort, jsonify
import random

app = Flask(__name__)

def is_admin_request():
    ip_headers = [
        'X-Forwarded-For',
        'X-Real-IP',
        'Client-IP',
        'X-Client-IP',
        'X-Custom-IP-Authorization'
    ]
    
    value = request.headers.get(random.choice(ip_headers))
    if value and '127.0.0.1' in value:
        return True

    if request.remote_addr == '127.0.0.1':
        return True
        
    return False

@app.route('/')
def index():
    return "<h1>Public endpoint</h1>"


@app.route('/secret')
def admin():
    if not is_admin_request():
        abort(403, description="Access Denied: Admin panel is restricted to localhost (127.0.0.1) only.")
        
    return "<h1>{LAB_PWNED}</h1>"

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
