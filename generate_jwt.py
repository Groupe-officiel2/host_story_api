import jwt
import datetime
import sys


SECRET_KEY = "secret"

def generate_jwt(user_id, role="user"):
    """
    Génère un token JWT pour un utilisateur donné.

    :param user_id: ID de l'utilisateur
    :return: Token JWT signé
    """
    payload = {
        "sub": user_id,
        "role": role,
        "exp": datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(hours=24)
    }

    token = jwt.encode(payload, SECRET_KEY, algorithm="HS256")
    return token

if __name__ == "__main__":
    user_id = "test_user"
    role = "admin"

    if len(sys.argv) > 1:
        user_id = sys.argv[1]
    if len(sys.argv) > 2:
        role = sys.argv[2]

    token = generate_jwt(user_id, role)
    print(f"Generated JWT Token: {token}")