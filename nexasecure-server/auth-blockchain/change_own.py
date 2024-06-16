class AuthenticationSystem:
    def __init__(self, dot_pool):
        self.dot_pool = dot_pool

    def authenticate(self, user, dot_id):
        for dot in self.dot_pool.dots:
            if dot.id == dot_id and dot.owner == user:
                return True
        return False
