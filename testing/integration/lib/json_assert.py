from itertools import zip_longest
from json_delta import udiff

from functools import partial
import json
dumps = partial(json.dumps, ensure_ascii=False, indent=2)


def assert_json_contains(expected, actual, msg="JSON mismatch"):
    """
    This method asserts that `expected` is a subset of `actual`
    parameters:
        @expected:
        @actual:
        @msg:
    """
    __tracebackhide__ = True

    subset = _extract_subset(actual, expected)
    if subset != expected:
        diff = '\n'.join(udiff(expected, subset, indent=2))
        raise AssertionError("{msg}:\n{diff}".format(msg=msg, diff=diff))


def _extract_subset(source, mask):
    if isinstance(mask, dict) and isinstance(source, dict):
        return _extract_dict_subset(source, mask)
    elif isinstance(mask, list) and isinstance(source, list):
        return _extract_list_subset(source, mask)
    else:
        return source


def _extract_dict_subset(source, mask):
    target = {}
    for key in mask:
        if key in source:
            target[key] = _extract_subset(source[key], mask[key])
    return target


MISSING = object()
def _extract_list_subset(source, mask):
    target = []
    for source_item, mask_item in zip_longest(source, mask, fillvalue=MISSING):
        if mask_item is MISSING:
            target.append(source_item)
        elif source_item is not MISSING:
            target.append(_extract_subset(source_item, mask_item))
    return target
