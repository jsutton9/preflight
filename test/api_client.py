import json
import requests

class Client:
    def __init__(self, target):
        self.target = target
        self.token = ""

    def add_user(self, email, password):
        token = ""
        with open("~/preflight/secret", "r") as f:
            token = f.read()
        url = self.target + "/users?token=" + token
        body = {"email": email, "password": password}
        response = requests.post(url, json.dumps(body))
        return response.content

    def authorize(self, email, password, permissions):
        req = {"permissions": permissions, 
                "expiry-hours": 24, 
                "description": "api test client"}
        url = self.target + "/tokens"
        response = requests.post(url, json.dumps(req), auth=(email, password))
        return json.loads(response.content)

    def change_password(self, newPassword):
        url = self.target + "/password?token=" + self.token
        response = requests.post(url, newPassword)

    def invoke_checklist(self, name):
        url = "%s/checklists/%s/invoke?token=%s" % (self.target, name, self.token)
        response = requests.post(url, "")

    def add_checklist(self, name, checklist):
        url = self.target + "/checklists?token=" + self.token
        req = {"name": name,
                "checklist": checklist}
        response = requests.post(url, json.dumps(req))

    def update_checklist(self, name, checklist):
        url = "%s/checklists/%s?token=%s" % (self.target, name, self.token)
        response = requests.put(url, json.dumps(checklist))

    def delete_checklist(self, name):
        url = "%s/checklists/%s?token=%s" % (self.target, name, self.token)
        response = requests.delete(url)

    def get_checklists(self):
        url = self.target + "/checklists?token=" + self.token
        response = requests.get(url)
        return json.loads(response.content)

    def update_global_setting(self, name, value):
        url = "%s/settings/%s?token=%s" % (self.target, name, self.token)
        response = requests.put(url, str(value))

    def get_global_settings(self):
        url = self.target + "/settings?token=" + self.token
        response = requests.get(url)
        return json.loads(response.content)

    def force_update(self):
        url = self.target + "/force-update?token=" + self.token
        response = requests.post(url, "")
