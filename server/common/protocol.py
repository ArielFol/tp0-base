import struct

from .utils import Bet
import datetime

def recv_bytes(sock, n):
    data = b''
    while len(data) < n:
        packet = sock.recv(n - len(data))
        if not packet:
            raise ConnectionError("socket connection lost")
        data += packet
    return data

def decode_bet(sock) -> Bet:
    agency = struct.unpack('>I', recv_bytes(sock, 4))[0]

    name_length = struct.unpack('>I', recv_bytes(sock, 4))[0]
    name = recv_bytes(sock, name_length).decode('utf-8')

    last_name_length = struct.unpack('>I', recv_bytes(sock, 4))[0]
    last_name = recv_bytes(sock, last_name_length).decode('utf-8')

    dni = struct.unpack('>Q', recv_bytes(sock, 8))[0]
    birthdate_no_format = struct.unpack('>q', recv_bytes(sock, 8))[0]
    number = struct.unpack('>I', recv_bytes(sock, 4))[0]

    birthdate = datetime.date.fromtimestamp(birthdate_no_format).isoformat()

    return Bet(agency, name, last_name, str(dni), birthdate, number)

def decode_bets_batch(sock) -> list[Bet]:
    bets = []
    bets_amount = struct.unpack('>I', recv_bytes(sock, 4))[0]
    for _ in range(bets_amount):
        bet = decode_bet(sock)
        bets.append(bet)

    if bets_amount != len(bets):
        return bets, ValueError(f"Expected {bets_amount} bets but received {len(bets)}")
        
    return bets, None
