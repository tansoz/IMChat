var electron = require('electron');
const net = require('net');

var BW = electron.BrowserWindow;
var App = electron.app;

App.allowRendererProcessReuse = false;
var SOCKET = new net.Socket();
var WIN = null;

function loadLoginBox () {
    // 创建浏览器窗口
    WIN = new BW({
        width: 310,
        height: 430,
        webPreferences: {
            nodeIntegration: true,
        },
        frame:false,
        resizable:false,
    });
    // 加载index.html文件
    WIN.loadFile('./template/Login/Login.html');
}

function loadMainBox(){

    WIN = new BW({
        width: 860,
        height: 700,
        webPreferences: {
            nodeIntegration: true,
        },
        frame:false,
        resizable:true,
    });

    WIN.loadFile('./template/Main/Main.html');

    SOCKET.write("GETUSERLIST:\r\n");

    // WIN.webContents.openDevTools();
}

var USERLIST = [];

electron.ipcMain.on('closeLoginBox',function(){
    WIN.close();
});
electron.ipcMain.on('minBox',function(){
    WIN.minimize();
});
electron.ipcMain.on('maxBox',function(){
    WIN.maximize();
});
electron.ipcMain.on('unmaxBox',function(){
    WIN.unmaximize();
});
electron.ipcMain.on('USERLIST',function(){
    WIN.webContents.send('USERLIST',{users:USERLIST});
});
electron.ipcMain.on('sendMsg',function(e,d){
    SOCKET.write(d.msg);
});
electron.ipcMain.on('WHOAMI',function(e,d){
    WIN.webContents.send('WHOAMI',{username:username});
});
electron.ipcMain.on('MISSMSG',function(e,d){
    SOCKET.write("GETMISSMSG:\r\n");
});
var username = "";
electron.ipcMain.on('clickLoginBtn',function(e,args){
    username = args.username.trim();
    if(username!=""){
        SOCKET.connect(4258,"127.0.0.1",function(){
            SOCKET.write("HELO:"+username+"\r\n");
        });
        SOCKET.on('data',function(e){
            var raw_s = e.toString().split("\r\n");

            for(var i = 0;i < raw_s.length - 1;i++){
                var s = raw_s[i] + "\r\n";
                if(s=="AUTH_SUCCEED\r\n"){
                    var loginbox = WIN;
                    loadMainBox();
                    loginbox.close();
                }else if(/^USERLIST:([^\r]+)\r\n$/.test(s)){
                    USERLIST = s.trim().substring(9).split(",");
                    WIN.webContents.send('USERLIST',{users:USERLIST});
                }else if(/^[^#@]+#[^@]+@[^\r]+\r\n$/.test(s)){
                    console.log(s);
                    var arr = s.match(/^([^#@]+)#([^@]+)@([^\r]+)\r\n$/);
                    WIN.webContents.send('NEWMSG',arr);
                }else if(/^REFRESHUSERLIST:\r\n$/.test(s)){
                    SOCKET.write("GETUSERLIST:\r\n");
                }
            }

        });
        SOCKET.on('close',function(e){
            var mainbox = WIN;
            loadLoginBox();
            mainbox.close();
            SOCKET = new net.Socket();
        });
        SOCKET.on('error',function(e){
            var mainbox = WIN;
            loadLoginBox();
            mainbox.close();
            SOCKET = new net.Socket();
        });
    }
});

App.whenReady().then(loadLoginBox); // 显示图形界面
