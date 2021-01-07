import sys
sys.path.append('.')

from rest.encoder import JSONEncoder
from werkzeug.exceptions import Unauthorized
from flask_reverse_proxy_fix.middleware import ReverseProxyPrefixFix

import rest

import argparse


def main():
    parser = argparse.ArgumentParser(description='Rest api server options')    
    parser.add_argument('--port', nargs='?', help='port for service (default 8400)')
    args = parser.parse_args()
    
    port = 8400
    if args.port:
        port = args.port
    print("Rest server on port", port)
    
    # create rest server
    rest.flask_app.app.json_encoder = JSONEncoder
    rest.flask_app.add_api('openapi/openapi.yaml', arguments={'title': 'SafeScale api specifications', 'server_url': 'localhost:8080'}, pythonic_params=True)
          
    # run
    #flask_app.run(host="0.0.0.0", port=port, debug=True, processes=16)
    rest.flask_app.run(host="0.0.0.0", port=port, debug=True, threaded=True)

#     api = connexion.App(__name__, options={"swagger_url": ""}, arguments={'server_url': "localhost:8400"})
#      
#     api.add_api('openapi/openapi.yaml')
#      
#     api.app.config['REVERSE_PROXY_PATH'] = 'http://localhost:8080'
#     ReverseProxyPrefixFix(api.app)
# 
#     api.run(host='0.0.0.0', port=port, debug=debug, server=deployment_server)

if __name__ == '__main__':
    main()




