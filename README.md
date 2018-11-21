# VStore
VStore is a file based, secure and versionned KV store.
The underlying storage is a git repository. The data is encrypted and decrypted from files.

## Usage.
```
vstore <local password> <file path> [<json pointer>] [<value>]
vstore aUjk87kdv credentials/gmail /login john.doe@gmail.com
vstore aUjk87kdv credentials/gmail /password gke94dsFVs
vstore aUjk87kdv credentials/gmail 
> {"login":"john.doe@gmail.com","password":"gke94dsFVs"}
vstore aUjk87kdv credentials/gmail /login
> john.doe@gmail.com
vsotre aUjk87kdv credentials/gmail /login jane.doe@gmail.com
vstore aUjk87kdv credentials/gmail /login
> jane.doe@gmail.com
vstore reset
> Trying to delete: /Users/john/Library/Caches/vstore
> Successfully deleted: /Users/john/Library/Caches/vstore
```

## Details.
On the first invocation, VStore ask for a master password and a remote repository.  The password is used to encrypt the data and the change in content get pushed to the remote for backup. This information is stored in a settings file encrypted with the local password.

VStore supports fuzzy matching of file path. If multiple or no file path match the input, VStore give the option to select one of them or to create a new one.


## Demo.
TODO

