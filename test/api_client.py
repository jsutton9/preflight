import json

import requests

class Client:
    def __init__(self, target, verify=False):
        self.target = target
        self.verify = verify
        self.token = ""

    def add_user(self, email, password):
        token = "wrong-token"
        #with open("~/preflight/secret", "r") as f:
        #    token = f.read().strip()
        url = self.target + "/users?token=" + token
        body = {"email": email, "password": password}
        response = requests.post(url, json.dumps(body), verify=self.verify)
        response.raise_for_status()
        return response.content

    def authorize(self, email, password, permissions):
        req = {"permissions": permissions, 
                "expiryHours": 24, 
                "description": "api test client"}
        url = self.target + "/tokens"
        response = requests.post(url, json.dumps(req), auth=(email, password), verify=self.verify)
        response.raise_for_status()
        self.token = response.json()["secret"]
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
