from flask import Flask, Response

app = Flask(__name__)

# Definované XML pro odpověď
rss_feed = '''<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
<title>Odstávky externích služeb SUPIN s.r.o</title>
<link>http://www.ckp.cz/Aplikace/Support/RSS/</link>
<description> | Plánované odstávky | Mimořádné odstávky | </description>
<lastBuildDate>Fri, 06 Sep 2024 09:34:19 +0200</lastBuildDate>
<language>cs</language>
<item>
<title>Plánovaná odstávka od 09.09.2024 09:00 do 11.09.2024 18:30</title>
<description>Aktualizace SVII (update 291 a 292)<h5>Služby mimo provoz:</h5><ol><li>X1 – Webová služba SVIPO II (Test)</li></ol></description>
<guid isPermaLink="false">66dab07b5452ss</guid>
<pubDate>Fri, 06 Sep 2024 09:34:19 +0200</pubDate>
</item>
</channel>
</rss>
'''

json_api = '''{"status": "ok"}
'''

# Endpoint, který vrací XML
@app.route('/rss', methods=['GET'])
def get_rss_feed():
    return Response(rss_feed, mimetype='application/xml')

# Endpoint, který vrací JSON
@app.route('/status', methods=['GET'])
def get_json_feed():
    return Response(json_api, mimetype='application/json')

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5050)
