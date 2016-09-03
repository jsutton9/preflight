import random

import requests

import api_client

c = api_client.Client("https://localhost")

try:
    email = "api-test%d@foo.bar" % random.randint(1, 9999)
    userId = c.add_user(email, "pass")
    print userId

    print c.authorize(email, "pass", {"checklistRead": True, "checklistWrite": True, "checklistInvoke": True, "generalRead": True, "generalWrite": True})
    #print c.authorize("api-test2@foo.bar", "pass", {"checklistInvoke": True})

    #print c.delete_checklist("bar")
    #print c.add_checklist("bar", {"tasksSource": "preflight", "tasksTarget": "preflight", "isScheduled": False, "tasks": ["x", "y"]})

    #print c.get_checklist("bar")

    #c.invoke_checklist("bar")

    #c.update_global_setting("timezone", "America/Denver")
    #print c.get_global_settings()

    print c.add_token({"checklistInvoke": True})

    tokens = c.get_tokens()
    print tokens
    token_id = tokens[1]["id"]

    print c.delete_token(token_id)
    print c.get_tokens()

    print c.delete_user(userId)

except requests.exceptions.HTTPError as e:
    print ""
    print "HTTPError from %s: " % e.request.url
    print e.response.status_code
    print e.response.content
    print ""
