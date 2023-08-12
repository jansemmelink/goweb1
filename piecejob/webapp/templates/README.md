# Templates #

Base templates:
* `page.tmpl` defines the generic layout for all pages

Other files:
* All pages (e.g. `login.tmpl`) must define head and body templates to be used in page.tmpl
* login-modal currently not used - shows how a popup form can be built...
* login-email-form asks email and when submitted, loginEmailHandler sends an OTP then shows login-otp
* login-otp-form accepts the OTP and verify