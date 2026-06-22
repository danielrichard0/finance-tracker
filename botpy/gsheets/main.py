from typing import Any
from google.oauth2 import service_account
from googleapiclient.discovery import build
from datetime import datetime, date
import pandas as pd

class ExpenseTrackerBaseClass:
    def __init__(self, sheetID: str):
        self.spreadsheet_id = sheetID
        self.scopes = ["https://www.googleapis.com/auth/spreadsheets"]
        self.creds = service_account.Credentials.from_service_account_file(
            "credentials.json",
            scopes=self.scopes,
        )
        self.service: Any = build("sheets", "v4", credentials=self.creds)
        self.sheet: Any = self.service.spreadsheets()
        self.date_format = '%d/%m/%Y'
        self.date_format2 = '%B %Y'
        self.data = []
        self.headers = ['No', 'Jenis', 'Pengeluaran', 'Tanggal', 'Qty']


    def _find_position_bydate(self, dates: list[str], date_append: date)->int:
        for i, date_ in enumerate(dates):
            if date_ == '' or date_== None:
                continue 
            try:        
                date_ = datetime.strptime(date_, '%d/%m/%Y').date()
            except ValueError:
                continue    
            
            if i > 0:
                if date_append < date_:                        
                    return i
                
        return len(dates)    


    def _get_values_from_columns(self, sheet_name:str ,cell1: str = 'A', cell2: str = 'A'):
        result = self.sheet.values().get(
            spreadsheetId=self.spreadsheet_id,
            range=f"{sheet_name}!{cell1}:{cell2}",
            majorDimension="COLUMNS"
        ).execute()

        return result.get("values")

    def _list_sheets(self):
        response = self.service.spreadsheets().get(
            spreadsheetId=self.spreadsheet_id,
            fields="sheets.properties(sheetId,title,index)"
        ).execute()

        val = {}
        for sheet in response.get("sheets", []):
            val[sheet['properties']['title']] = sheet['properties']['sheetId']      
        
        return val      
    

    def _get_all_data(self):
        self.data = self._get_values_from_columns('FORMAT','F', 'J')
        dframe = {}

        if len(self.headers) != len(self.data):
            raise IndexError
        
        # the header len needs to be the same as self.data len
        for a, header in enumerate(self.headers):
            dframe[header] = self.data[a][:len(self.data[0])]

        self.data = pd.DataFrame(dframe)
        print(self.data)


    def _parse_date(self, date: str):
        date_ = datetime.strptime(date, self.date_format).date()   
        return date_.strftime(self.date_format2 ) 

    def make_new_sheet(self, sheet_name):
        request = {
                "addSheet": {
                    "properties" : {
                        "title" : sheet_name
                    }
                    
                }
            }
        
        a = self.sheet.batchUpdate(spreadsheetId=self.spreadsheet_id, body={"requests" : request}).execute()
        print(a)


    def format_row(self, range: list[list[int], list[int]], option: str = "LEFT"):        
        requests = [ 
            {
                "repeatCell": {
                    "range": {
                        "sheetId": self.spreadsheet_id,
                        "startRowIndex": 1,   # 0-based, row 2
                        "endRowIndex": 2,
                        "startColumnIndex": 0, # column A
                        "endColumnIndex": 1,
                    },
                    "cell": {
                        "userEnteredFormat": {
                            "horizontalAlignment": option  # LEFT | CENTER | RIGHT
                        }
                    },
                    "fields": "userEnteredFormat.horizontalAlignment"
                }        
            }
        ]   

        req = self.sheet.batchUpdate(spreadsheetId=self.spreadsheet_id, body={"requests": requests})
        req.execute()



    def insert_row(self, sheet_name: str, row_index: int, values: list, column: str = "A"):
        """
        Insert a blank row at `row_index` (0-based) in the sheet named `sheet_name`
        and write `values` across the columns starting at column A.

        Example: `insert_row('Sheet1', 4, ['a', 'b', 'c'])` inserts before the
        current row 5 and writes the values into A5:C5.
        """


        sheet_id = self._list_sheets().get(sheet_name, None)

        if sheet_id is None:
            raise ValueError(f"Sheet not found: {sheet_name}")

        requests = [
            {
                "insertDimension": {
                    "range": {
                        "sheetId": sheet_id,
                        "dimension": "ROWS",
                        "startIndex": row_index,
                        "endIndex": row_index + 1,
                    },
                    "inheritFromBefore": False,

                }                
            },
            {
                "repeatCell": {
                    "range": {
                        "sheetId": sheet_id,
                        "startRowIndex": row_index,   # 0-based, row 2
                        "endRowIndex": row_index + 1,
                        "startColumnIndex": 6, # column A
                        "endColumnIndex": 11,
                    },
                    "cell": {
                        "userEnteredFormat": {
                            "horizontalAlignment": 'CENTER'  # LEFT | CENTER | RIGHT
                        }
                    },
                    "fields": "userEnteredFormat.horizontalAlignment"
                }        
            }            
        ]

        self.service.spreadsheets().batchUpdate(spreadsheetId=self.spreadsheet_id, body={"requests": requests}).execute()

        range_ = f"{sheet_name}!{column}{row_index + 1}"
        body = {"values": [values]}
        self.service.spreadsheets().values().update(
            spreadsheetId=self.spreadsheet_id,
            range=range_,
            valueInputOption="RAW",
            body=body,
        ).execute()


    def insert_row_by_date(self, values: list):  
        sheet_name = self._parse_date(values[3])
        sheet_id = self._list_sheets().get(sheet_name, None)
        if sheet_id == None:
            self.make_new_sheet(sheet_name)

        return     

        date = datetime.strptime(values[3], self.date_format).date()
        dates = self._get_values_from_columns('FORMAT','I', 'I')[0]
        pos = self._find_position_bydate(dates=dates, date_append=date)        

        if pos <= 0:
            raise IndexError
        else:
            self.insert_row(sheet_name, pos, values, 'F')   


    def insert_sheet():
        return     
    

if __name__ == '__main__':
    Base = ExpenseTrackerBaseClass('19bOnjPALH7TbSeyJr8N11KiIpCpMBxFBSxvtv6ViYss')
    #result = Base.make_new_sheet('TESTING EDP')
    result = Base.insert_row_by_date(['1', 'Kebutuhan', 'Testing dari API', '29/6/2026', '350000'])

    #result = Base._get_all_data()
    #result = Base.insert_row_by_date('FORMAT', ['1', 'Kebutuhan', 'Testing dari API', '29/4/2025', '350000'])