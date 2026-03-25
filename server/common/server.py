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
        self._bets_lock = threading.Lock()
        self._finished_lock = threading.Lock()
        self._sorteo_event = threading.Event()


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
        threads = []

        while not self._shutdown_event.is_set():
            try:
                client_sock = self.__accept_new_connection()
                client_thread = threading.Thread(target=self.__handle_client_connection, args=(client_sock,))
                client_thread.start()
                threads.append(client_thread)
            except socket.timeout:
                continue

        logging.info("action: joining_threads | result: in_progress")
        for thread in threads:
            thread.join()
        logging.info("action: joining_threads | result: success")

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
                handle_bets_message(client_sock, self._bets_lock)

            elif message_type == MessageType.FIN:
                handle_finish_message(client_sock, self._finished_agencies, self._finished_lock, self._sorteo_event)

            elif message_type == MessageType.RESULTS:
                handle_results_message(client_sock, self._sorteo_event, self._bets_lock)

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
