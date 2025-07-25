# gaia

## Modules

### CRM
Handles user data

Handles authentication with MitID 
- Handle MitID access token
- Handle MitID user info
- Handle matching between MitID identity and Gaia identity
- Issue "access token" with aud and scope 

## Auth
Responsibility:

Offer an authentication service.
- Provide authentication from external identity provider MitId
- Provide onboarding 
- Provide reverse proxy for additional calls
- Use this as a logging point


This could be an independent service where 
- services are registered with scope, aud and role
- match between mitiduuid and gaia user is made
    - APP -> JWT -> CRM match -> JWT -> APP


Only build if there is a point in it.


### DMI 
Handles "consumption" data

### App 
Handles the frontend 


## Design

Authentication flow happens with an external identity provider (MitID) that returns JWT tokens

Authorization flow happens with CRM 