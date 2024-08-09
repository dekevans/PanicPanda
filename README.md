# INSTRUCTIONS TO RUN:

1.  Go must be installed on your machine. If not installed, follow this site's instructions to install: [(https://go.dev/doc/install)](https://go.dev/doc/install)
2.  To run the script (starting from fuzz-swag directory):
<pre><code>go . run</code></pre>
There are two ways to enter the required data:
a. Manually, through the cli
<pre><code>Input domain and paths (everything before the paths on the swaggerdoc):
Input auth token (if none, leave blank):
Input timer:
Input wordlist file path: (if you want pure random data, leave blank)
Is your swaggerdoc v2 or v3?
Input swagger file path:
Do you want to fuzz the headers? (Y/N)   
How many seconds do you want to wait before retrying the fuzzer after continuous failure?
</code></pre>
b. Piping it through the cli (config.txt in this example contains the information normally entered through the cli in a text file, separated by new lines) 
<pre><code>go . run < config.txt</code></pre>

### There is an example config.txt included, fill it out, don't forget to add a new line at the end
### There is also an example wordlist.txt included, to fuzz portions of the words with {}, add valid path names after PATHLIST, as shown in wordlist.txt example
# RATIONALE

- Currently, the code takes in swagger documentation and parses the api paths such that every endpoint gets fuzzed
- The code starts a new thread for every path that it finds, and fuzzes it for the specified length of time
- The code uses the google go fuzzer and wordlist as the poet
- The code replaces certain parts of the wordlist and uses others as a seed for randomness
- The code uses the go http request sender as the courier, using the results of the poet to fill out the http request
- The code uses the response sent back by backend and analyzes it's length, expected codes, and errors, and decides if it needs to be sent back into the poet as a corpus (if there were already existing corpuses in the poet).

<img width="1594" alt="Screenshot 2024-08-09 at 2 28 25 PM" src="https://github.com/user-attachments/assets/7f9cfe9a-7d8e-44ea-9753-8ff1472a55d5">


  
# GUIDANCE REQUESTED

- Not DoSing the backends I'm testing
    - Thought: using semphores (7/15/2024)
- Overall correctness of concept
    - Still needs advice 

# TO-DO

- Rate-Limiting system

# CREATED BY: RITE (Bryant H) Summer 2024
