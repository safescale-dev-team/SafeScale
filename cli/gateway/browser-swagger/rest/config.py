import yaml

class Config(object):
    
    def __init__(self):
        self.config = {} 
        
    def load(self, file):
        import os
        print(os.getcwd())
        
        with open(file) as content:
            self.config = yaml.full_load(content)
    