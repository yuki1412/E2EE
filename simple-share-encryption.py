import random

class Person:
    def __init__(self, name):
        self.name = name
        self.private_key = random.randint(0, 100)
        self.public_key = None
        self.shared_secret = None

    def generate_public_key(self, baseNumber):
        self.public_key = (baseNumber ** self.private_key) % 100

    def generate_shared_secret(self, other_public_key):
        self.shared_secret = (other_public_key ** self.private_key) % 100

def main():
    baseNumber = random.randint(1, 100)
    print(f"Base number: {baseNumber}")
    alice = Person("Alice")
    bob = Person("Bob")

    print(f"Generating public keys")
    alice.generate_public_key(baseNumber)
    bob.generate_public_key(baseNumber)

    print(f"Key travelled through the internet: Alice: {alice.public_key}, Bob: {bob.public_key}")
    alice.generate_shared_secret(bob.public_key)
    bob.generate_shared_secret(alice.public_key)

    print(f"Shared secret for {alice.name}: {alice.shared_secret}")
    print(f"Shared secret for {bob.name}: {bob.shared_secret}")

if __name__ == "__main__":
    main()  