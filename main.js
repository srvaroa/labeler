const spawn = require('child_process').spawn;
const actionAssetUrl = getInput('action_asset_url')

function getInput(name) {
    const value = process.env[`INPUT_${name.toUpperCase()}`]
    
    if (!value) {
        throw new Error(`Cannot find input ${name}`)
    }

    return value
}

async function sh(command) {
    return new Promise((resolve, reject) => {
        const cmd  = spawn('sh', ['-c', command]);

        cmd.stdout.on('data', function(data) {
            console.log(data.toString());
        });
        
        cmd.stderr.on('data', function(data) {
            console.error(data.toString());
        });
        
        cmd.on('exit', function(code) {
            if (code > 0) {
                reject(code);
            } else {
                resolve(code);
            }
        });
    })
}

async function run() {
    await sh(`curl -sSL ${actionAssetUrl} | tar -xvzf-`)
    await sh('./action')
}

run()
.then(() => {
    console.log('Done')
})
.catch((errCode) => {
    console.error(`Failed with error code: ${errCode}`)
})
