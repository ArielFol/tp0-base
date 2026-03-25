import logging
import os


from .protocol import decode_bets_batch, decode_agency_id, encode_no_results_message, encode_results_message
from .utils import get_winners_for_agency, store_bets

def handle_bets_message(client_sock, bets_lock):
    bets, err = decode_bets_batch(client_sock)

    with bets_lock:
        store_bets(bets)

    if err is not None:
        logging.info(f'action: apuesta_recibida | result: fail | cantidad: {len(bets) if bets is not None else 0}')
        client_sock.sendall(b'400\n')
        return

    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets) if bets is not None else 0}')
    client_sock.sendall(b'200\n')

def handle_finish_message(client_sock, finished_agencies, finished_lock, sorteo_event):

    agency_id, err = decode_agency_id(client_sock)
    if err is not None:
        logging.info(f'action: mensaje_recibido | result: fail | error: no se pudo decodificar el id de agencia')
        client_sock.sendall(b'400\n')
        return

    with finished_lock:
        if agency_id in finished_agencies:
            logging.info(f'action: mensaje_recibido | result: fail | error: agencia {agency_id} ya habia sido finalizada')
            client_sock.sendall(b'400\n')
            return
        finished_agencies.add(agency_id)
        if len(finished_agencies) == int(os.getenv("AGENCIES_AMOUNT", 0)):
            logging.info(f'action: sorteo | result: success')
            sorteo_event.set()

    logging.info(f'action: mensaje_recibido | result: success | tipo: mensaje de finalizacion')
    client_sock.sendall(b'200\n')



def handle_results_message(client_sock, sorteo_event, bets_lock):
    logging.info(f'action: mensaje_recibido | result: success | tipo: mensaje de resultados')

    agency_id, err = decode_agency_id(client_sock)
    if err is not None:
        logging.info(f'action: mensaje_recibido | result: fail | error: no se pudo decodificar el id de agencia')
        encoded_message = encode_no_results_message()
        client_sock.sendall(encoded_message)
        return
    
    if not sorteo_event.is_set():
        encoded_message = encode_no_results_message()
        client_sock.sendall(encoded_message)
        return

    winners = get_winners_for_agency(agency_id, bets_lock)

    encoded_message = encode_results_message(winners)
    client_sock.sendall(encoded_message)
    logging.info(f'action: mensaje_enviado | result: success | tipo: mensaje de resultados | cantidad_ganadores: {len(winners)}')
