# Progress #
- added on_enter_actions which can call functions, getMySkills() returns hard coded list
- switched to redis session store because sqlite limited size and ran out
- added list display for skills and jobs
    - loads list of items when display/refresh
    - item has caption
    - list supports operations at the bottom
    - list.item selection opens the item
    - tabular display is working needed
- added edit
    - get works without id, but busy figuring out how to pass args to get func
    - save works without id - should anyway get id from internal
    - cancel not yet working to go back

# Busy With #
- editormodule to get/edit/save with user struct types from get/set func
    editor can now edit profile,
    editor must also be able to work on list selection when Item is already defined by list.
        but item then is map... not struct.

- editor must get with id, e.g. profile[id]
    with two concurrent windows, not yet separated

- editor also need option to view not in form or view in form but read only all fields
    and enable/disable edit when nothing changed
    and implement validation rules with active javascript feedback (may be in react)

- start with action, e.g.
    if nat_id is not set, go to ask it, else load profile
    if profile valid, go home, else ask nat id...

- handle [Cancel] from editor, may be javascript to do something?

- consider using jq instead of template to extract session values for func req etc...?
    and pass value as interface{} always then func can assert it has required type and extract fields as needed

- add conditional logic in actions, e.g.
    {"if":{"expr":"len(NationalId) != 13", "then":[], "else":[]}}

# Bugs #
- When from menu click on something that fail, then something that works, click browser back to menu, then all clicks on menu go to last thing that worked... something wierd...

- session data keeps growing, e.g. when list is displayed again and again, more uuid added... need way to scope and cleanup page specific data including links and Item<uuid> must also be temporary

# Todo #
- list displays (not menu)
    - change list order
    - filter on part of item values/caption
    - limit display needed for long lists, show total matches

- profile page must show back if nothing changes, else show save/cancel
    - make this also a generic item to edit a struct or map with some constraints built into the struct type, returned from the load func and also need a save func.

    ...i.e. generic list and generic view/display/delete/add given a table name and item type

- app custom display modules, like list and menu and prompt... but allow app to register own modules, need to register them as item types, instead of hard coded item struct at moment... see how action was done.

- move templates into app to be generic and let use change them
- add default style sheet and extend with custom style sheets

- try not to build a back-end yet... will be useful to make react app later, but for
    now the aim is exactly the opposit, i.e. to make an app quickly standalone
    and later be able to call other services as needed, but initially just all-in-one
    quick to market.
- make user id accessible to functions, e.g. national id to call backend
- template resolve inside action function args and set...

- indicate unused JSON attributes when parsing the app file to avoid surprises and things not being applied

- let session expire and inform user and implement user register/login/auth, but still allow display of some pages without auth and indicate when auth is needed

- imply and show back option on all menus but give option to disable it
- all session data must be stacked between item nav so that back will drop some values and restore old values
- pass named params to user functions - not directly retrieve session data
    - so function can iterated over lists etc...

- forms with multiple fields
- radio input
- checkbox input
- integration and back-end

- need to be persist app version to continue on the same version if new one is being rolled out
- need to be able to include sub apps - each with own version too
- need to be able to track ongoing apps on each version

- dynamic generated menu items must be stored in the session with unique id
    when selected, the id can be looked up in session data and link to next item

- need to register and install identified menu items at the start so can jump on item id
    load items from JSON with implementation in code

- ability to run without user interface to run long running workflows as well in the same framework

- merge multiple JSON files and go packages into one app for modularity

- extend to cover long running processes by calling external services as consumer/producers

