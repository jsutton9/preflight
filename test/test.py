import api_client

c = api_client.Client("https://localhost")

#print c.add_user("api-test2@foo.bar", "pass")

print c.authorize("api-test2@foo.bar", "pass", {"checklistRead": True, "checklistWrite": True, "checklistInvoke": True})

#print c.delete_checklist("bar")
#print c.add_checklist("bar", {"tasksSource": "preflight", "tasksTarget": "preflight", "isScheduled": False, "tasks": ["x", "y", "z"]})

#print c.get_checklist("bar")

c.invoke_checklist("bar")
