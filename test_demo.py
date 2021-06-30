import json
import time
import typing
import uuid

import requests

schema = 'http'
api_host = 'localhost:8082'
proxy_host = 'localhost:8080'

api_url = f'{schema}://{api_host}'
proxy_url = f'{schema}://{proxy_host}'


def create_service(service_host: str, base_rest_url: str, base_ws_url: str):
    url = f'{api_url}/service/'
    payload = {
        'host': service_host,
        'base_rest_url': base_rest_url,
        'base_ws_url': base_ws_url,
        'desc': 'service desc',
        'logo': 'https://via.placeholder.com/150?text=Apron',
        'create_time': int(time.time()),
        'service_provider_name': 'sp',
        'service_provider_account': 'sp_account',
        'service_usage': 'usage',
        'service_price_plan': 'price',
        'service_declaimer': 'declaimer',
    }
    r = requests.post(url, json=payload)
    assert r.status_code == 201


def create_key(service_id: str) -> str:
    url = f'{api_url}/service/{service_id}/keys/'
    print(url)
    r = requests.post(url, json={'account_id': 'foobar'})
    print(r.content)
    assert r.status_code == 200

    rslt = r.json()
    assert rslt['serviceId'] == service_id
    return rslt['key']


def send_request(user_key: str, payload: typing.Dict):
    url = f'{proxy_url}/v1/{user_key}/anything/aaa'
    r = requests.post(url, json=payload)
    assert r.status_code == 200
    print(json.dumps(r.json(), indent=2))


def fetch_usage_report():
    url = f'{api_url}/service/report/'
    r = requests.get(url)
    print(r.content)
    assert r.status_code == 200
    print(json.dumps(r.json(), indent=2))


if __name__ == '__main__':
    service_host = proxy_host
    base_rest_url = 'https://httpbin.org/'
    base_ws_url = 'ws://localhost:8765/'
    create_service(service_host, base_rest_url, base_ws_url)
    k = create_key(service_host)

    for _ in range(3):
        print('-' * 20)
        send_request(k, {'now': time.time()})
        time.sleep(1)

    print('+' * 20)
    fetch_usage_report()
