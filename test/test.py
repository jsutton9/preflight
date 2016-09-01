import api_client

c = api_client.Client("https://localhost")

print c.add_user("api-test3@foo.bar", "pass")

print c.authorize("api-test2@foo.bar", "pass", {"checklistRead": True, "checklistWrite": True, "checklistInvoke": True, "GeneralRead": True, "GeneralWrite": True})

print c.delete_checklist("bar")
print c.add_checklist("bar", {"tasksSource": "preflight", "tasksTarget": "preflight", "isScheduled": False, "tasks": ["x", "y"]})

print c.get_checklist("bar")

#c.invoke_checklist("bar")

#c.update_global_setting("timezone", "America/Chicage")
#print c.get_global_settings()
