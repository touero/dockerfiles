#!/usr/bin/env python3
import http.server
import socketserver
import os
from html import escape

os.chdir("./")
PORT = 80

class MyHandler(http.server.SimpleHTTPRequestHandler):
    def list_directory(self, path):
        try:
            entries = os.listdir(path)
        except OSError:
            self.send_error(404, "No permission to list directory")
            return None
        entries.sort(key=lambda a: a.lower())

        displaypath = escape(self.path)
        self.send_response(200)
        self.send_header("Content-type", "text/html; charset=utf-8")
        self.end_headers()

        self.wfile.write(
            b"""<!DOCTYPE html>
                <html>
                    <head>
                    <meta charset='utf-8'>
                    <title>Directory listing</title>
                    <style>
                        body { background-color: #282828; color: #ebdbb2; font-family: sans-serif; }
                        a { color: #83c07c; text-decoration: none; }
                        a:hover { color: #fabd2f; }
                        hr { border-color: #928374; }
                        ul { list-style-type: none; padding-left: 0; }
                        li { padding: 2px 0; }
                        p { color: #928374; }
                        div.footer { text-align:center; color:#928374; margin-top:10px; }
                    </style>
                    </head>
                    <body>"""
        )

        self.wfile.write(f"<h2>Directory listing for {displaypath}</h2><hr><ul>".encode())

        for name in entries:
            fullname = os.path.join(path, name)
            displayname = linkname = name
            if os.path.isdir(fullname):
                displayname += "/"
                linkname += "/"

            color = "#83a598" if os.path.isdir(fullname) else "#ebdbb2"
            self.wfile.write(f'<li><a href="{escape(linkname)}" style="color:{color}">{escape(displayname)}</a></li>'.encode())

        self.wfile.write(b"</ul><hr>")

        self.wfile.write(b'<p>These executables are provided by <a href="https://github.com/touero" style="color:#83a598;">touero</a>.</p>')
        self.wfile.write(
            b"""<p>Licensed under the Apache License, Version 2.0 (the "License");
                you may not use this file except in compliance with the License.
                You may obtain a copy of the License at
                <a href="http://www.apache.org/licenses/LICENSE-2.0">http://www.apache.org/licenses/LICENSE-2.0</a>.
                Unless required by applicable law or agreed to in writing, software
                distributed under the License is distributed on an "AS IS" BASIS,
                WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
                See the License for the specific language governing permissions and
                limitations under the License.</p>"""
        )
        self.wfile.write(
            b"""<p>See <a href="https://github.com/touero/dockerfiles">https://github.com/touero/dockerfiles</a>
                for more information and our source code.</p>"""
        )

        self.wfile.write(b"</body></html>")

        return None

with socketserver.TCPServer(("", PORT), MyHandler) as httpd:
    print(f"Serving at http://localhost:{PORT}")
    httpd.serve_forever()
