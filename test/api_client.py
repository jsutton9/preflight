import json

import requests

class Client:
    def __init__(self, target, verify=False, secret_file="/etc/preflight/secret"):
        self.target = target
        self.verify = verify
        self.token = ""
        self.email = ""
        self.password = ""
        with open(secret_file, "r") as f:
            self.node_secret = f.read().strip()

    def add_user(self, email, password):
        url = self.target + "/users?nodeSecret=" + self.node_secret
        body = {"email": email, "password": password}
        response = requests.post(url, json.dumps(body), verify=self.verify)
        response.raise_for_status()
        return response.content

    def delete_user(self, userId):
        url = self.target + "/users/%s?nodeSecret=%s" % (userId, self.node_secret)
        response = requests.delete(url, verify=self.verify)
        response.raise_for_status()

    def authorize(self, email, password, permissions):
        self.email = email
        self.password = password
        token = self.add_token(permissions)
        self.token = token["secret"]
        return self.token

    def change_password(self, newPassword):
        url = self.target + "/password?token=" + self.token
        response = requests.post(url, newPassword, verify=self.verify)
        response.raise_for_status()

    def invoke_checklist(self, name):
        url = "%s/checklists/%s/invoke?token=%s" % (self.target, name, self.token)
        response = requests.post(url, "", verify=self.verify)
        response.raise_for_status()

    def add_checklist(self, name, checklist):
        url = self.target + "/checklists?token=" + self.token
        req = {"name": name,
                "checklist": checklist}
        response = requests.post(url, json.dumps(req), verify=self.verify)
        response.raise_for_status()
        return response.headers["Location"]

    def update_checklist(self, name, checklist):
        url = "%s/checklists/%s?token=%s" % (self.target, name, self.token)
        response = requests.put(url, json.dumps(checklist), verify=self.verify)
        response.raise_for_status()

    def delete_checklist(self, name):
        url = "%s/checklists/%s?token=%s" % (self.target, name, self.token)
        response = requests.delete(url, verify=self.verify)
        response.raise_for_status()

    def get_checklists(self):
        url = self.target + "/checklists?token=" + self.token
        response = requests.get(url, verify=self.verify)
        response.raise_for_status()
        return response.json()

    def get_checklist(self, name):
        url = "%s/checklists/%s?token=%s" % (self.target, name, self.token)
        response = requests.get(url, verify=self.verify)
        response.raise_for_status()
        return response.json()

    def update_global_setting(self, name, value):
        url = "%s/settings/%s?token=%s" % (self.target, name, self.token)
        response = requests.put(url, str(value), verify=self.verify)
        response.raise_for_status()

    def get_global_settings(self):
        url = self.target + "/settings?token=" + self.token
        response = requests.get(url, verify=self.verify)
        response.raise_for_status()
        return response.json()

    def force_update(self):
        url = self.target + "/force-update?token=" + self.token
        response = requests.post(url, "", verify=self.verify)
        response.raise_for_status()

    def get_tokens(self):
        url = self.target + "/tokens?token=" + self.token
        response = requests.get(url, verify=self.verify)
        response.raise_for_status()
        return response.json()

    def add_token(self, permissions):
        req = {"permissions": permissions,
                "expiryHours": 24,
                "description": "api test client"}
        url = self.target + "/tokens"
        response = requests.post(url, json.dumps(req), auth=(self.email, self.password), verify=self.verify)
        response.raise_for_status()
        return response.json()

    def delete_token(self, token_id):
        url = self.target + "/tokens/%s?token=%s" % (token_id, self.token)
        response = requests.delete(url, verify=self.verify)
        response.raise_for_status()
