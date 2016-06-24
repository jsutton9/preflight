import json
import requests

class Client:
    def __init__(self, target):
        self.target = target
        self.token = ""

    def add_user(self, email, password):
        # TODO
        pass

    def authorize(self, email, password, permissions):
        # TODO
        pass

    def change_password(self, newPassword):
        # TODO
        pass

    def invoke_checklist(self, name):
        # TODO
        pass

    def add_checklist(self, name, checklist):
        # TODO
        pass

    def update_checklist(self, name, checklist):
        # TODO
        pass

    def delete_checklist(self, name):
        # TODO
        pass

    def get_checklists(self):
        # TODO
        pass

    def update_global_setting(self, name, value):
        # TODO
        pass

    def get_global_settings(self):
        # TODO
        pass

    def force_update(self):
        # TODO
        pass

    def set_todoist_token(self, token):
        # TODO
        pass

    def set_trello_token(self, token):
        # TODO
        pass
