# Design #
- i18n and substitution on all captions
- prompts
- imply and show back option on all menus but give option to disable it
- work with session values, set, clear, retrieve, ...
- add actions
- call registered user functions as actions
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

