import sys
import os


## load configuration
from rest.config import Config
config = Config()
config.load('rest/config.yaml')

# create Flask application
print("[init]", "flask")
import connexion
flask_app = connexion.App(__name__)
