# gaia



## Modules

### CRM
Handles user data

Handles authentication with MitID 
- Handle MitID access token
- Handle MitID user info
- Handle matching between MitID identity and Gaia identity
- Issue "access token" with aud and scope 



### DMI 
Handles "consumption" data

### App 
Handles the frontend 


## Design

Authentication flow happens with an external identity provider (MitID) that returns JWT tokens

Authorization flow happens with CRM 