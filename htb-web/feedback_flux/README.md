# Feedback Flux
Solves [https://app.hackthebox.com/challenges/Feedback Flux](https://app.hackthebox.com/challenges/Feedback%20Flux)

## Exploit 
XSS / CSRF to steal flag from localStorage via [https://github.com/TYPO3/html-sanitizer/security/advisories/GHSA-mm79-jhqm-9j54](https://github.com/TYPO3/html-sanitizer/security/advisories/GHSA-mm79-jhqm-9j54)

## Arguments
`target_url`: URL (IP + Port) of the HackTheBox instance
`bin_url`: URL of the callback endpoint to receive the exfiltrated flag
