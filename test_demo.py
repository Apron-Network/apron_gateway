import json
import time
import typing
import uuid

import requests

api_url = 'http://localhost:8082'
proxy_url = 'http://localhost:8080'


def create_service(service_id, service_name: str, base_url: str, schema: str):
    url = f'{api_url}/service/'
    payload = {
        'id': service_id,
        'name': service_name,
        'base_url': base_url,
        'schema': schema,
    }
    r = requests.post(url, json=payload)
    assert r.status_code == 201


def create_key(service_id: str) -> str:
    url = f'{api_url}/service/{service_id}/keys/'
    r = requests.post(url, json={'account_id': 'foobar'})
    assert r.status_code == 200

    rslt = r.json()
    assert rslt['serviceId'] == service_id
    return rslt['key']


def send_request(service_id: str, user_key: str, payload: typing.Dict):
    url = f'{proxy_url}/v1/{service_id}/{user_key}/anything/aaa'
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
    service_id = uuid.uuid4().hex
    service_name = str(int(time.time() * 1e3))
    base_url = 'localhost:2345/'
    schema = 'http'
    create_service(service_id, service_name, base_url, schema)
    k = create_key(service_id)

    for _ in range(3):
        print('-' * 20)
        send_request(service_id, k, {'now': time.time()})
        time.sleep(1)

    print('+' * 20)
    fetch_usage_report()
