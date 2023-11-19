import g4f

g4f.debug.logging = True  # Enable logging
g4f.check_version = False  # Disable automatic version checking
print(g4f.version)  # Check version
print(g4f.Provider.Ails.params)  # Supported args

# Automatic selection of provider
content = "Thanks for reaching out. I'm just passing my 4 year anniversary at current company and starting to look at new opportunities. I've heard of Layer Zero and definitely an exciting company. Do you still look for engineers?"
# Streamed completion
response = g4f.ChatCompletion.create(
    model="gpt-3.5-turbo",
    messages=[{"role": "user", "content": "rewrite this in natural english speaker: " + content}],
    # stream=True,
)

to_reply = """Hi Bob,

Just wanted to reach out one last time so you didn't miss this exciting opportunity to join the Backend team at a $3 billion valuation startup! I won't reach out anymore after this as I have already emailed you a few times, if you're not interested let me know! Otherwise I'll try to check back again in a few months :)


Best,

Kenneth
Recruiting @ LayerZero Labs"""
params = "i want to take this opportunity"
response = g4f.ChatCompletion.create(
    model="gpt-3.5-turbo",
    messages=[{"role": "user", "content": "reply this email that " + params + ' the email is: '+ to_reply}],
    # stream=True,
)

print(response)

# for message in response:
#     print(message, flush=True, end='')

# # Normal response
# response = g4f.ChatCompletion.create(
#     model=g4f.models.gpt_4,
#     messages=[{"role": "user", "content": "Hello"}],
# )  # Alternative model setting

# print(response)