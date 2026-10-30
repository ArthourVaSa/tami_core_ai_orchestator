class Settings:
    def __init__(self):
        self.active_model = "llama3.1"
        self.port = 3500

# Esta variable global será compartida entre la consola interactiva y el servidor web
global_settings = Settings()