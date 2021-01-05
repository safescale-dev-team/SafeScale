
def to_utf8_str(value):
    if type(value) is bytes:
        return value.decode('utf-8')
    return value
