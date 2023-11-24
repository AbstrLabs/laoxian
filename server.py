import g4f
import zmq
import os
import sys
from string import Template
import logging
import json
import traceback
from concurrent.futures import ThreadPoolExecutor, as_completed


logging.basicConfig(level=logging.INFO)

g4f.debug.logging = True  # Enable logging
g4f.check_version = False  # Disable automatic version checking
logging.info('gp4free version: ' + g4f.version)  # Check version
logging.info(g4f.Provider.Ails.params)  # Supported args


def parse_message(message):
    try:
        payload = json.loads(message)
    except json.decoder.JSONDecodeError:
        return {'error': 'invalid json'}, None, None
    try:
        template = payload['template']
        params = payload['params']        
    except KeyError:
        return {'error': 'invalid payload'}, None, None
    return None, template, params


def find_template(template_name):
    script_directory = os.path.dirname(os.path.abspath(sys.argv[0]))
    template_file = os.path.join(script_directory, 'prompts', template_name+'.txt')
    if os.path.isfile(template_file):
        with open(template_file) as f:
            template = Template(f.read())
        return None, template
    else:
        return {'error': 'template does not exist'}, None

def make_prompt(template, params):
    try:
        print(template)
        print(params)
        return None, template.substitute(params)
    except KeyError:
        return {'error': 'template param missing'}, None

def process_request(message):
    error, template_name, params = parse_message(message)
    if error:
        return error
    error, template = find_template(template_name)
    if error:
        return error
    error, prompt = make_prompt(template, params)
    if error:
        return error
    error, completion = generate_completion(prompt)
    if error:
        return error
    return {'completion': completion}
    
usable = [
    g4f.Provider.ChatBase,
    g4f.Provider.Chatgpt4Online,
    g4f.Provider.GPTalk,
    g4f.Provider.GeekGpt,
    g4f.Provider.GptForLove,
    g4f.Provider.You
]

def gpt_responder(msg):
    def responder(provider):
        return g4f.ChatCompletion.create(
            model="gpt-3.5-turbo",
            provider=provider,
            messages=[{"role": "user", "content": msg}],
            proxy="http://localhost:8889"
        )
    return responder

def generate_completion(prompt):
    with ThreadPoolExecutor() as executor:
        for ret in executor.map(gpt_responder(prompt), usable):
            if isinstance(ret, Exception):
                logging.error('generated an exception: %s' % ret)
            else:
                return None, ret
    return {"error": "all provider response error"}, None

context = zmq.Context()
socket = context.socket(zmq.REP)
logging.info(f"Current libzmq version is {zmq.zmq_version()}")
logging.info(f"Current  pyzmq version is {zmq.__version__}")
socket.bind("tcp://*:5555")
logging.info('zmq server running at tcp://127.0.0.1:5555')

while True:
    #  Wait for next request from client
    message = socket.recv()
    print(f"Received request: {message}")
    try:
        #  Do some 'work'
        response = process_request(message)
    except Exception:
        traceback.print_exc(file=sys.stderr)
        response = {"error": "internal error"}
    #  Send reply back to client
    socket.send_string(json.dumps(response))