# access_key_manager
CLI for rotating AWS Access Keys

## Install

Using Homebrew:

    brew install stuarthicks/tap/access_key_manager

Using Go:

    go install github.com/stuarthicks/access_key_manager@latest

## Usage

```
# access_key_manager -h   
Usage of access_key_manager:
  -delete string
    	Delete an access key
  -list
    	List access keys
  -rotate string
    	Rotate an access key
```

### Summary

```
access_key_manager -list
access_key_manager -rotate AKIAOLDKEY
access_key_manager -delete AKIAOLDKEY
```

### Complete Steps

1. First list your access keys.

```
# access_key_manager -list
Access Key ID: AKIASIGDOORANVN4MFOO
Creation Date: 2022-11-02 13:37:26 +0000 UTC
Status: Active
```

2. Start the rotation (creates a new access key and marks the old one as inactive).
```
# access_key_manager -rotate AKIASIGDOORANVN4MFOO
2022/11/16 16:22:55 created new access key
2022/11/16 16:22:55 marked access key "AKIASIGDOORANVN4MFOO" as inactive

New credentials
---------------
Access Key ID: AKIASIGDOORALNRZNBAR
Secret Access Key: hunter2
Creation Date: 2022-11-16 16:22:56 +0000 UTC
Status: Active

Found AWS Credentials file at: ~/.aws/credentials
Automatically updated credentials file.

After confirming the new credentials work, use -delete [ID] to delete the previous access key
```

3. If the automatic credentials file updated failed, manually update your aws credentials file for the new access key based on the output from `-rotate`.

4. Delete the old/inactive access key.
```
# access_key_manager -delete AKIASIGDOORANVN4MFOO
2022/11/16 16:23:50 successfully deleted access key "AKIASIGDOORANVN4MFOO"
```

5. (Optional) Delete the credentials file backup at `~/.aws/credentials.bak`
