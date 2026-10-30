import os
import threading
import sys
import requests
import uvicorn
from fastapi import FastAPI
from app.presentation.api import router
from app.config import global_settings

from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.markdown import Markdown
from rich.text import Text
from rich.align import Align
from prompt_toolkit import PromptSession
from prompt_toolkit.completion import WordCompleter
from prompt_toolkit.styles import Style

app = FastAPI(title="TAMI Core", version="1.0.0")
app.include_router(router)
console = Console()

def run_server():
    uvicorn.run(app, host="0.0.0.0", port=global_settings.port, log_level="error")

def display_welcome_banner():
    """Renderiza una pantalla de bienvenida idéntica a la estética retro-terminal en color TAMI Mint."""
    console.clear()
    
    # 1. Caja de bienvenida superior izquierda (Color TAMI)
    welcome_text = Text("* Welcome to TAMI Core", style="bold #b4ffeb")
    console.print(Panel.fit(welcome_text, border_style="#b4ffeb"))
    console.print("\n")
    
    # 2. Subtítulo centrado
    subtitle = Text("TAMI Core VS Standard Proxies", style="bold white")
    console.print(Align.center(subtitle))
    console.print("\n")
    
    # 3. Arte ASCII gigante estilo bloques
    ascii_tami = """
████████╗ █████╗ ███╗   ███╗██╗
╚══██╔══╝██╔══██╗████╗ ████║██║
   ██║   ███████║██╔████╔██║██║
   ██║   ██╔══██║██║╚██╔╝██║██║
   ██║   ██║  ██║██║ ╚═╝ ██║██║
   ╚═╝   ╚═╝  ╚═╝╚═╝     ╚═╝╚═╝
    """
    
    # Agrupamos el ASCII gigante y el subtítulo lateral
    layout = Table.grid(padding=4)
    layout.add_column(justify="right")
    layout.add_column(justify="left", vertical="middle")
    
    layout.add_row(
        Text(ascii_tami, style="bold #b4ffeb"), 
        Text("CORE\n\nThe local AI orchestrator agent", style="dim white")
    )
    
    console.print(Align.center(layout))
    console.print("\n\n")
    
    # 4. Pausa de continuación
    console.print("[dim white]Press Enter to continue...[/]")
    input()
    console.clear()

def chat_with_ai(user_input: str, current_model: str):
    try:
        with console.status(f"[bold #b4ffeb]TAMI ({current_model}) está pensando...[/]", spinner="dots"):
            response = requests.post(
                f"http://localhost:{global_settings.port}/v1/chat/completions",
                json={
                    "model": current_model,
                    "messages": [{"role": "user", "content": user_input}]
                },
                timeout=60
            )
            response.raise_for_status()
            data = response.json()
            ai_response = data["choices"][0]["message"]["content"]
            
        console.print(Panel(Markdown(ai_response), border_style="#b4ffeb", title="🤖 Respuesta", title_align="left"))
        
    except Exception as e:
        console.print(f"[bold red]❌ Error al comunicarse con la IA:[/] {str(e)}")

def start_interactive_cli():
    server_thread = threading.Thread(target=run_server, daemon=True)
    server_thread.start()

    # Mostrar la pantalla inicial al arrancar
    display_welcome_banner()

    # Encabezado persistente de la terminal
    console.print(f"🌐 [bold green]Proxy Activo:[/] http://localhost:{global_settings.port}")
    console.print("💡 [dim]Escribe para chatear. Usa '/' para comandos (ej. /help)[/]\n")

    comandos_base = ['/model', '/list', '/status', '/clear', '/exit', '/help']
    completer = WordCompleter(comandos_base, ignore_case=True)
    
    # La magia de los colores en el input
    style = Style.from_dict({
        'prompt': 'bold #b4ffeb', # El verde menta principal de TAMI
        'model': 'bold #73ffda'   # Un turquesa un poco más profundo para el nombre del modelo
    })
    
    session = PromptSession(completer=completer, style=style)

    while True:
        try:
            try:
                real_status = requests.get(f"http://localhost:{global_settings.port}/admin/config/status").json()
                current_model = real_status['active_model']
            except:
                current_model = "desconectado"

            prompt_text = [
                ('class:prompt', 'TAMI '),
                ('class:model', f'({current_model})'),
                ('class:prompt', ' > ')
            ]
            
            cmd = session.prompt(prompt_text).strip()

            os.set_blocking(sys.stdout.fileno(), True)

            if not cmd: continue
                
            if cmd.startswith('/'):
                parts = cmd.split()
                command = parts[0].lower()

                if command in ["/exit", "/quit"]:
                    console.print("[bold red]Apagando TAMI Core... ¡Hasta pronto![/]")
                    sys.exit(0)
                    
                elif command == "/help":
                    table = Table(show_header=True, header_style="bold magenta")
                    table.add_column("Comando", style="#b4ffeb")
                    table.add_column("Descripción")
                    table.add_row("/model <nombre>", "Cambia el motor LLM (ej. /model qwen2.5:7b)")
                    table.add_row("/list", "Muestra los modelos locales instalados")
                    table.add_row("/status", "Muestra el estado del Proxy")
                    table.add_row("/clear", "Limpia la terminal")
                    table.add_row("/exit", "Cierra el orquestador")
                    console.print(table)
                    
                elif command == "/list":
                    try:
                        ollama_tags = requests.get("http://localhost:11434/api/tags").json()
                        table = Table(title="📦 Modelos Locales Disponibles", title_style="bold yellow")
                        table.add_column("Modelo", style="cyan")
                        table.add_column("Tamaño", justify="right", style="green")
                        for m in ollama_tags.get("models", []):
                            table.add_row(m['name'], m['details'].get('parameter_size', 'N/A'))
                        console.print(table)
                    except:
                        console.print("[bold red]❌ Error:[/] Ollama no está corriendo.")

                elif command == "/model":
                    if len(parts) > 1:
                        nuevo_modelo = parts[1]
                        try:
                            ollama_tags = requests.get("http://localhost:11434/api/tags").json()
                            modelos_instalados = [m['name'] for m in ollama_tags.get("models", [])]
                            
                            if nuevo_modelo not in modelos_instalados and f"{nuevo_modelo}:latest" not in modelos_instalados:
                                console.print(f"[bold red]❌ Error:[/] El modelo '{nuevo_modelo}' no está instalado en Ollama.")
                                continue
                                
                            requests.post(f"http://localhost:{global_settings.port}/admin/config/model", json={"model_name": nuevo_modelo})
                            console.print(f"[bold green]✅ Cerebro actualizado exitosamente a {nuevo_modelo}[/]")
                        except requests.exceptions.ConnectionError:
                            console.print("[bold red]❌ Error al contactar al servidor.[/]")
                    else:
                        console.print("[bold red]❌ Faltó el modelo.[/] Usa: /model qwen2.5:7b")
                        
                elif command == "/status":
                    console.print(Panel(f"🔌 [bold]Puerto:[/] {global_settings.port}\n🧠 [bold]Modelo:[/] [cyan]{current_model}[/]", title="Estado", border_style="green"))
                    
                elif command == "/clear":
                    console.clear()
                    
                else:
                    console.print(f"[bold red]❌ Comando '{command}' no reconocido.[/] Usa '/help'.")
            
            else:
                chat_with_ai(cmd, current_model)
                
        except KeyboardInterrupt:
            continue
        except EOFError:
            console.print("\n[bold red]Apagando TAMI Core...[/]")
            sys.exit(0)

if __name__ == "__main__":
    start_interactive_cli()