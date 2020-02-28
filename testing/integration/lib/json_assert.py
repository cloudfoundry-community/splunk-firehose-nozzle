
def assert_json_contains(expected, actual, msg="JSON mismatch"):
    """
    Asserts that `expected` is a subset of `actual`
    """
    # TODO: build this function to compare json content
    #  `expected` and `actual` are meant to be JSON serializable data structures

    if actual != expected:
        raise AssertionError("{msg}".format(msg=msg))


