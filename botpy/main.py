import discord
import os
from dotenv import load_dotenv
from discord import app_commands
import aiohttp
from enum import Enum
from gsheets.main import ExpenseTrackerBaseClass

load_dotenv()
TOKEN = os.getenv('DISCORD_TOKEN')
MY_GUILD = discord.Object(id=os.getenv('DISCORD_GUILD_ID'))

class MyClient(discord.Client):
    def __init__(self, *, intents, **options):
        super().__init__(intents=intents, **options)
        self.tree = app_commands.CommandTree(self)
        self.BaseSheetClass = ExpenseTrackerBaseClass(sheetID='19bOnjPALH7TbSeyJr8N11KiIpCpMBxFBSxvtv6ViYss')

    async def on_ready(self):
        print(f'Logged on as {self.user}!')

    async def on_message(self, message):
        print(f'Message from {message.author}: {message.content}')

    async def setup_hook(self):
        # This copies the global commands over to your guild.
        self.tree.copy_global_to(guild=MY_GUILD)
        await self.tree.sync(guild=MY_GUILD)

class Categories(Enum):
    Jajan = 'JAJAN'
    Bulanan = 'BULANAN'
    Makan = 'MAKAN'
    ETC = 'ETC'

class Types(Enum):
    Expense= 'E'
    Income= 'I'    

intents = discord.Intents.default()
intents.message_content = True

client = MyClient(intents=intents)

@client.tree.command(name="help", description="List all the commands that are available in this bot")
async def help(interaction: discord.Interaction):
    await interaction.response.send_message("""
    /help - list all the commands that are available in this bot
    /add_transaction - add transaction to your account                                            
    """)

@client.tree.command(name="hello", description='katakan hello')
async def hello(interaction: discord.Interaction):
    await interaction.response.send_message('ini masuk')

@client.tree.command(description='Tambah transaksi baru')    
async def add_transaction(interaction: discord.Interaction,
                          title: str, amount: app_commands.Range[float, 0, None],
                          category: Categories, notes: str | None, transaction_date: str, type: Types ):
    
    url='http://127.0.0.1:8080/api/v1/transactions'
    payload = {
        "title": title,
        "amount" : amount,
        "category": category.value, 
        "notes": notes,
        "transaction_date": transaction_date,
        "type": type.value
    }

    try :
        async with aiohttp.ClientSession() as session:        
            async with session.post(
                url,
                json=payload
            ) as resp:
                data = await resp.json()
    except aiohttp.ClientConnectionError as e:
        await interaction.response.send_message('Server did not response after awhile') 
        return False               

    await interaction.response.send_message(f"Created transaction: {data}")

@client.tree.command(description='Tambah transaksi baru2')    
async def add_transaction2(interaction: discord.Interaction,
                          title: str, amount: app_commands.Range[float, 0, None],
                          category: Categories, notes: str | None, transaction_date: str, type: Types ):
    
    payload = {
        "title": title,
        "amount" : amount,
        "category": category.value, 
        "notes": notes,
        "transaction_date": transaction_date,
        "type": type.value
    }

    client.BaseSheetClass.testing()            
    
    await interaction.response.send_message(f"Created transaction:")
client.run(TOKEN)