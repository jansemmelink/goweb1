# Progress #
- added on_enter_actions which can call functions, getMySkills() returns hard coded list

# Busy With #
- list display (not menu)
    - caption template can display []string or fields from []struct
    - change sort order
    - filter on part of caption
    - limit display
    - list operations like add new item, or send to all, etc...
    need ability to load list each time rendered

# Todo #
- dynamic display modules built in the app...

- try not to build a back-end yet... will be useful to make react app later, but for
    now the aim is exactly the opposit, i.e. to make an app quickly standalone
    and later be able to call other services as needed, but initially just all-in-one
    quick to market.
- make user id accessible to functions, e.g. national id to call backend
- template resolve inside action function args and set...

- imply and show back option on all menus but give option to disable it
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

