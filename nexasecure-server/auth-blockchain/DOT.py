import uuid

class DOT:
    def __init__ (self,owner,metadata) -> None :
        self.id = str(uuid.uuid4())
        self.owner = owner
        self.metadata = metadata

    def transfer(self,new_owner):
        self.owner = new_owner


class DOT_Pool:
    def __init__(self) -> None:
        self.pool = []
    
    def mint_DOT(self,owner,metadata):
        dot = DOT(owner,metadata)
        self.pool.append(dot)
        return dot
    
    def transfer_DOT(self,dot,new_owner):
        for DOT in self.pool:
            if DOT.id == dot.id:
                DOT.transfer(new_owner)
                self.pool.remove(DOT)
                return DOT
        return False
