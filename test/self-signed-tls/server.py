import http.server
import ssl

class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/":
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.end_headers()
            self.wfile.write(b"OK")
        else:
            self.send_response(404)
            self.end_headers()

server = http.server.HTTPServer(("0.0.0.0", 4443), Handler)

ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
ctx.load_cert_chain("localhost.localdomain/tls.crt", "localhost.localdomain/tls.key")
server.socket = ctx.wrap_socket(server.socket, server_side=True)

print("Serving on https://localhost:4443")
server.serve_forever()
