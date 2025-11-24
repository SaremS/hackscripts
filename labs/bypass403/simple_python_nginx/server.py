from flask import Flask, request, abort

app = Flask(__name__)

@app.route('/')
def home():
    return "<h1>Public Area</h1><p>Try to access /secret</p>"

@app.route('/<path:resource>')
def catch_all(resource):
    normalized_resource = resource.lower()
    
    if normalized_resource == 'secret':
        return "<h1>{LAB_PWNED}</h1>"
    
    return f"<h1>Not Found</h1><p>Resource '{resource}' does not exist.</p>", 404

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
