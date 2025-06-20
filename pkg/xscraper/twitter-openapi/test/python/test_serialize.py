import base64
import concurrent.futures
import glob
import inspect
import json
import logging
import os
import time
import traceback
import warnings
from enum import Enum
from pathlib import Path
from typing import Any

import bs4
import openapi_client as pt
import requests
import urllib3
from x_client_transaction import ClientTransaction
from x_client_transaction.utils import (
    generate_headers,
    get_ondemand_file_url,
    handle_x_migration,
)

warnings.filterwarnings("ignore")
logging.basicConfig(level=logging.DEBUG, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger("test_serialize")

TWITTER_SESSION = os.environ.get("TWITTER_SESSION", None)
ERROR_UNCATCHED = os.environ.get("ERROR_UNCATCHED", "false").lower() == "true"
SLEEP_TIME = float(os.environ.get("SLEEP", "0"))
CUESOR_TEST_COUNT = int(os.environ.get("CUESOR_TEST_COUNT", "3"))
STRICT_MODE = os.environ.get("STRICT_MODE", "false").lower() == "true"
MULTI_THREAD = os.environ.get("MULTI_THREAD", "true").lower() == "true"


def get_key(snake_str):
    components = snake_str.split("_")
    return "".join(x.title() for x in components[1:])


def get_cursor(obj, fn):
    res = []
    if isinstance(obj, dict):
        callback = fn(obj)
        if callback is not None:
            res.extend(callback)
        else:
            for v in obj.values():
                res.extend(get_cursor(v, fn))
    elif isinstance(obj, list):
        for v in obj:
            res.extend(get_cursor(v, fn))
    return res


def find_cursor(x):
    if x.get("__typename") is pt.TypeName.TIMELINETIMELINECURSOR:
        return [x["value"]]


def find_name(x):
    if x.get("name") is not None:
        return [x["name"]]


def get_transaction_base():
    session = requests.Session()
    session.headers = generate_headers()
    home_page_response = handle_x_migration(session=session)
    home_page = session.get(url="https://x.com")
    home_page_response = bs4.BeautifulSoup(home_page.content, "html.parser")
    ondemand_file_url = get_ondemand_file_url(response=home_page_response)
    ondemand_file = session.get(url=ondemand_file_url)
    ondemand_file_response = bs4.BeautifulSoup(ondemand_file.content, "html.parser")
    ct = ClientTransaction(home_page_response, ondemand_file_response)
    return ct


def get_transaction_id(key, ct=get_transaction_base()):
    return ct.generate_transaction_id(
        method=placeholder[key]["@method"], path=placeholder[key]["@path"]
    )


def get_kwargs(key, additional):
    kwargs = {"path_query_id": placeholder[key]["queryId"], "_headers": {}}
    if placeholder[key].get("variables") is not None:
        kwargs["variables"] = json.dumps(placeholder[key]["variables"] | additional)
    if placeholder[key].get("features") is not None:
        kwargs["features"] = json.dumps(placeholder[key]["features"])
    if placeholder[key].get("fieldToggles") is not None:
        kwargs["field_toggles"] = json.dumps(placeholder[key]["fieldToggles"])
    if placeholder[key].get("@path") is not None:
        kwargs["_headers"]["x-client-transaction-id"] = get_transaction_id(key)
    return kwargs


def match_rate_zero(a, b, base, key):
    def get(obj, key):
        if isinstance(obj, list):
            return get(obj[key[0]], key[1:])
        if obj.__dict__.get("actual_instance") is not None:
            return get(obj.actual_instance, key)
        if len(key) == 0:
            return obj
        return get(super_get(obj.__dict__, key[0]), key[1:])

    if STRICT_MODE:
        obj_name = type(get(base, key[:-1]))
        obj_key = f"{obj_name.__name__}.{key[-1]}"
        raise Exception(f"Not defined: {obj_key}\nContents: {b}")

    return 0


def match_rate(a, b, base, key=""):
    if isinstance(a, Enum):
        a = a.value
    if isinstance(b, Enum):
        b = b.value
    if a is None and b is False:
        return 1
    if a is False and b is None:
        return 1
    if a is None and isinstance(b, list) and len(b) == 0:
        return 1
    if isinstance(a, list) and b is None and len(a) == 0:
        return 1
    if a is None and isinstance(b, dict) and len(b) == 0:
        return 1
    if isinstance(a, dict) and b is None and len(a) == 0:
        return 1
    if isinstance(a, dict) and isinstance(b, dict):
        if len(a) == 0 and len(b) == 0:
            return 1
        marge_key = set(a.keys()) | set(b.keys())
        data = [match_rate(a.get(k), b.get(k), base, [*key, k]) for k in marge_key]
        return sum(data) / len(b)
    if isinstance(a, list) and isinstance(b, list):
        if len(a) == 0 and len(b) == 0:
            return 1
        if len(a) != len(b):
            return match_rate_zero(a, b, base, key)
        data = [match_rate(a[i], b[i], base, [*key, i]) for i in range(len(a))]
        return sum(data) / len(a)
    if a == b:
        return 1
    return match_rate_zero(a, b, base, key)


def save_cache(data):
    rand = time.time_ns()
    os.makedirs("cache", exist_ok=True)
    with open(f"cache/{rand}.json", "w+") as f:
        json.dump(data, f, indent=4)


def super_get(obj: dict, key: str):
    keys = [
        key,
        "".join(["_" + c.lower() if c.isupper() else c for c in key]).lstrip("_"),
    ]

    for k in keys:
        if obj.get(k) is not None:
            return obj[k]
    raise KeyError(key)


def task_callback(file, thread=True):
    try:
        with open(file, "r") as f:
            cache = json.load(f)
        data = pt.__dict__[cache["type"]].from_json(cache["raw"])

        rate = match_rate(
            data.to_dict(),
            json.loads(cache["raw"]),
            base=data,
        )
        return rate, file
    except Exception:
        if thread:
            return 0, file
        else:
            raise


def kebab_to_upper_camel(text: dict[str, Any]) -> dict[str, Any]:
    res = {}
    for key, value in text.items():
        new_key = "".join([x.capitalize() for x in remove_prefix(key).split("-")])
        res[new_key] = value
    return res


def get_header(data: dict, name: str):
    ignore = ["host", "connection"]
    return {key: value for key, value in data[name].items() if key not in ignore}


def remove_prefix(text: str) -> str:
    if text.startswith("x-twitter-"):
        return text[10:]
    if text.startswith("x-"):
        return text[2:]
    return text


def error_dump(e):
    if ERROR_UNCATCHED:
        raise

    logger.error("==========[STACK TRACE]==========")
    for trace in traceback.format_exc().split("\n"):
        logger.error(trace)
    logger.info("================================")


if __name__ == "__main__":
    if Path("cookie.json").exists():
        with open("cookie.json", "r") as f:
            cookies = json.load(f)
    elif TWITTER_SESSION is not None:
        data = base64.b64decode(TWITTER_SESSION).decode("utf-8")
        cookies = json.loads(data)
    else:
        commands = ["python -m pip install tweepy_authlib", "python tools/login.py"]
        raise Exception(
            f"cookie.json not found. Please run `{'; '.join(commands)}` first."
        )

    if isinstance(cookies, list):
        cookies = {k["name"]: k["value"] for k in cookies}
    cookies_str = "; ".join([f"{k}={v}" for k, v in cookies.items()])

    with open("src/config/placeholder.json", "r") as f:
        placeholder = json.load(f)

    fail = []
    files = glob.glob("cache/*.json")
    if MULTI_THREAD:
        with concurrent.futures.ProcessPoolExecutor() as executor:
            tasks = [executor.submit(task_callback, x) for x in files]
            for task in concurrent.futures.as_completed(tasks):
                rate, file = task.result()
                if rate < 1:
                    fail.append(file)
                logger.info(f"Match rate: {rate}")
    else:
        for file in files:
            rate, file = task_callback(file, thread=False)
            if rate < 1:
                fail.append(file)
            logger.info(f"Match rate: {rate}")

    logger.info(f"Fail: {len(fail)} / {len(glob.glob('cache/*.json'))}")

    for file in fail:
        task_callback(file, thread=False)
        logger.info(f"Match rate: {rate}")

    for file in glob.glob("other/**/*.json", recursive=True):
        with open(file, "r") as f:
            data = json.load(f)

        try:
            _ = pt.TweetDetailResponse.from_dict(data)
        except Exception as e:
            error_dump(e)
    access_token = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
    twitter_url = "https://x.com/home"
    latest_user_agent_res = urllib3.PoolManager().request(
        "GET",
        "https://raw.githubusercontent.com/fa0311/latest-user-agent/refs/heads/main/header.json",
    )
    latest_user_agent = json.loads(latest_user_agent_res.data.decode("utf-8"))
    api_conf = pt.Configuration(
        api_key=kebab_to_upper_camel(
            {
                **get_header(latest_user_agent, "chrome-fetch"),
                "sec-ch-ua-platform": '"Windows"',
                "accept-encoding": "identity",
                "referer": twitter_url,
                "priority": "u=1, i",
                "authorization": f"Bearer {access_token}",
                "x-twitter-auth-type": "OAuth2Session",
                "x-twitter-client-language": "en",
                "x-twitter-active-user": "yes",
                "x-csrf-token": cookies["ct0"],
            },
        ),
    )
    api_conf.access_token = access_token
    api_client = pt.ApiClient(configuration=api_conf, cookie=cookies_str)
    api_client.user_agent = get_header(latest_user_agent, "chrome-fetch")["user-agent"]

    session = requests.Session()
    session.headers = get_header(latest_user_agent, "chrome")

    error_count = 0

    for x in [pt.DefaultApi, pt.TweetApi, pt.UserApi, pt.UsersApi, pt.UserListApi]:
        for props, fn in inspect.getmembers(x):
            if not callable(fn):
                continue
            if props.startswith("__") or not props.endswith("_with_http_info"):
                continue

            key = get_key(props[:-15])
            cursor_list = set([None])
            cursor_history = set()

            try:
                for _ in range(CUESOR_TEST_COUNT):
                    cursor = cursor_list.pop()
                    cursor_history.add(cursor)
                    logger.info(f"Try: {key} {cursor}")

                    kwargs = get_kwargs(
                        key, {} if cursor is None else {"cursor": cursor}
                    )
                    res: pt.ApiResponse = getattr(x(api_client), props)(**kwargs)
                    data = res.data.to_dict()

                    save_cache(
                        {
                            "raw": res.raw_data.decode("utf-8"),
                            "type": res.data.__class__.__name__,
                        }
                    )

                    new_cursor = set(get_cursor(data, find_cursor)) - cursor_history
                    cursor_list.update(new_cursor)

                    rate = match_rate(
                        data,
                        json.loads(res.raw_data),
                        res.data,
                    )
                    logger.info(f"Match rate: {rate}")

                    if data.get("errors") is not None:
                        logger.error(data)
                        error_count += 1

                    if len(cursor_list) == 0:
                        break
                    time.sleep(SLEEP_TIME)

            except Exception as e:
                error_dump(e)
                error_count += 1

    try:
        logger.info("Try: Self UserByScreenName Test")
        kwargs = get_kwargs("UserByScreenName", {"screen_name": "ptcpz3"})
        res = pt.UserApi(api_client).get_user_by_screen_name_with_http_info(**kwargs)
        data = res.data.to_dict()

        rate = match_rate(
            data,
            json.loads(res.raw_data),
            res.data,
        )
        logger.info(f"Match rate: {rate}")
        screen_name = data["data"]["user"]["result"]["legacy"]["screen_name"]
        if not screen_name == "ptcpz3":
            raise Exception("UserByScreenName failed")
    except Exception as e:
        error_dump(e)
        error_count += 1

    ids = [
        # "1180389371481976833", banned
        "900282258736545792",
        "1212617657003859968",
        "2455740283",
        "2326837940",
        "1788224200639160320",
        "1500128450186985472",
        "25073877",
    ]
    for id in ids:
        try:
            logger.info("Try: Self UserTweets Test")
            kwargs = get_kwargs("UserTweets", {"userId": id})
            res = pt.TweetApi(api_client).get_user_tweets_with_http_info(**kwargs)
            data = res.data.to_dict()

            rate = match_rate(
                data,
                json.loads(res.raw_data),
                res.data,
            )
            logger.info(f"Match rate: {rate}")

        except Exception as e:
            error_dump(e)
            error_count += 1

    ids = [
        "1720975693524377759",
        "1721006592303251551",
        "1739194269477331076",
        # "1697450269259522256",
        # "1697450278742884799",
        "1749500209061663043",
        "1759056048764469303",
        "1349129669258448897",
        "1810188416812019999",
        "1851981523207299417",
        "1853879226987901408",
        "1866022435549757577",
        "1866103697148887145",
        "1866036470303309863",
        "1862405433639804958",
        "1869983867521904840",
        "1875050002046726519",
        "1848219562136801480",
        "1881993128288399684",
        "1899104692577489182",
    ]
    for id in ids:
        try:
            logger.info(f"Try: Self TweetDetail {id} Test")
            kwargs = get_kwargs("TweetDetail", {"focalTweetId": id})
            res = pt.TweetApi(api_client).get_tweet_detail_with_http_info(**kwargs)
            data = res.data.to_dict()

            save_cache(
                {
                    "raw": res.raw_data.decode("utf-8"),
                    "type": res.data.__class__.__name__,
                }
            )

            rate = match_rate(
                data,
                json.loads(res.raw_data),
                res.data,
            )
            logger.info(f"Match rate: {rate}")
        except Exception as e:
            error_dump(e)
            error_count += 1

    if error_count > 0:
        exit(1)
