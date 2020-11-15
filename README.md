# mx-counter

Get a count of mail domains from an email list.

## Summary

Given

```
tom@tomontheinternet.com
jane@yahoo.ca
joe@gmail.com
fred@youtube.com
ingrid@slack.com
rory@microsoft.com
```

Outputs

```
google.com 4
yahoodns.net 1
outlook.com 1
```

## Usage

From a file:

`mx-counter emails.txt`

From stdin:

`cat emails.txt | mx-counter`:w
