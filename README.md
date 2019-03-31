# VStore
VStore is a file based, encrypted and versionned KV store.
The underlying storage is a git repository. The data is encrypted and decrypted from files.

## Usage.
```
VSTORE_PASSWORD=<local_password> vstore (get|set) <file path> [<json pointer>]
echo 'john.doe@gmail.com' | pbcopy
VSTORE_PASSWORD=aUjk87kdv vstore set credentials/gmail /login
echo gke94dsFVs | pbcopy
VSTORE_PASSWORD=aUjk87kdv vstore credentials/gmail /password
VSTORE_PASSWORD=aUjk87kdv vstore get credentials/gmail 
> {"login":"john.doe@gmail.com","password":"gke94dsFVs"}
VSTORE_PASSWORD=aUjk87kdv vstore get credentials/gmail /login
> john.doe@gmail.com
echo 'jane.doe@gmail.com' | pbcopy
VSTORE_PASSWORD=aUjk87kdv vstore set credentials/gmail /login
VSTORE_PASSWORD=aUjk87kdv vstore get credentials/gmail /login
> jane.doe@gmail.com
vstore reset
> Trying to delete: /Users/john/Library/Caches/vstore
> Successfully deleted: /Users/john/Library/Caches/vstore
```

## Details.
On the first invocation, VStore ask for a master password and a remote repository.  The password is used to encrypt the data and the change in content get pushed to the remote for backup. This information is stored in a settings file encrypted with the local password.

VStore supports fuzzy matching of file path. If multiple or no file path match the input, VStore give the option to select one of them or to create a new one.

## Disclaimer.
I'm not a security expert. Use at your own risk.

## Demo.
TODO

