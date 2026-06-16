import discord
import os
from dotenv import load_dotenv
from discord import app_commands
import aiohttp
from enum import Enum

load_dotenv()
TOKEN = os.getenv('DISCORD_TOKEN')
MY_GUILD = discord.Object(id=os.getenv('DISCORD_GUILD_ID'))

class MyClient(discord.Client):
    def __init__(self, *, intents, **options):
        super().__init__(intents=intents, **options)
        self.tree = app_commands.CommandTree(self)

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

@client.tree.command(name="hello", description='katakan hello')
async def hello(interaction: discord.Interaction):
    await interaction.response.send_message('ini masuk')

@client.tree.command(description='Tambah transaksi baru')    
async def add_transaction(interaction: discord.Interaction,
                          title: str, amount: app_commands.Range[float, 0, None],
                          category: Categories, notes: str, transaction_date: str, type: Types ):
    url='http://127.0.0.1:8080/api/v1/transactions'
    payload = {
        "title": title,
        "amount" : amount,
        "category": category.value, 
        "notes": notes,
        "transaction_date": transaction_date,
        "type": type.value
    }

    print('payload : ', interaction.user.id)

    async with aiohttp.ClientSession() as session:
        async with session.post(
            url,
            json=payload
        ) as resp:
            data = await resp.json()

    await interaction.response.send_message(
        f"Created transaction: {data}")

    # if resp.status_code==200:
    #     interaction.response.send_message('berhasil simpan transaksi!')
    # else:
    #     interaction.response.send_message('gagal simpant transaksi')     

client.run(TOKEN)