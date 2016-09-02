import requests

import api_client

c = api_client.Client("https://localhost")

try:
    #print c.add_user("api-test3@foo.bar", "pass")

    print c.authorize("api-test2@foo.bar", "pass", {"checklistRead": True, "checklistWrite": True, "checklistInvoke": True, "GeneralRead": True, "GeneralWrite": True})
    #print c.authorize("api-test2@foo.bar", "pass", {"checklistInvoke": True})

    #print c.delete_checklist("bar")
    #print c.add_checklist("bar", {"tasksSource": "preflight", "tasksTarget": "preflight", "isScheduled": False, "tasks": ["x", "y"]})

    #print c.get_checklist("bar")

    #c.invoke_checklist("bar")

    c.update_global_setting("timezone", "America/Denver")
    print c.get_global_settings()

except requests.exceptions.HTTPError as e:
    print ""
    print "HTTPError from %s: " % e.request.url
    print e.response.status_code
    print e.response.content
    print ""
