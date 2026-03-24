import os
import socket
import logging
import signal
import threading

from .protocol import decode_message_type
from .utils import store_bets
from .handlers import handle_bets_message, handle_finish_message, handle_results_message
from enum import IntEnum

class MessageType(IntEnum):
    BETS = 1
    FIN = 2
    RESULTS = 3

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._server_socket.settimeout(1)
        self._shutdown_event = threading.Event()
        self._finished_agencies = set()

        signal.signal(signal.SIGTERM, self._handle_sigterm)

    def _handle_sigterm(self, signum, frame):
        logging.info('action: shutdown_server | result: in_progress')
        self._shutdown_event.set()

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while not self._shutdown_event.is_set():
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except socket.timeout:
                continue
        
        logging.info("action: close_socket | result: in_progress")
        self._server_socket.close()
        logging.info("action: close_socket | result: success")
        logging.info("action: shutdown_server | result: success")

    

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            logging.info(f'action: decoding_message | result: in_progress')
            message_type = decode_message_type(client_sock)
            
            if message_type == MessageType.BETS:
                handle_bets_message(client_sock)

            elif message_type == MessageType.FIN:
                handle_finish_message(client_sock, self._finished_agencies)

            elif message_type == MessageType.RESULTS:
                #read environment variable to know how many agencies to expect
                total_agencies = int(os.getenv("AGENCIES_AMOUNT", 0))
                handle_results_message(client_sock, self._finished_agencies, total_agencies)

        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success')
        return c
