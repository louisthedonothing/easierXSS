



# EasyXSS 


## Overview

This XSS scanner is simple, easy to use, and entirely go-based.
It runs through a specified wordlist of common XSS vulnerabilities, and alerts when one is potentially found


## Features:
- Scans web pages for possible XSS vulnerabilities
- Injects a set of XSS scripts from a wordlist (One is provided in the repository)
- Identifies XSS vulnerabilities by inspecting server response from the script


## Prerequisites
- Python 3+ **MUST** be installed on the system
- The following python libraries **MUST** be installed:
    'Requests', for making http requests
    'bs4' for identifying input forms
    'ArgParse' for handling command line prompts
    
You can install the libraries by running:
```pip install requests bs4 Argparse```

## Usage

``` ./easyxss.go -u "http://example.com" -p "PATH TO WORDLIST"```

# Socials:

twitter: @looouuuiiiss_
Discord: gynxci
