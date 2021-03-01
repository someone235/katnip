const baseEndpoint = process.env.REACT_APP_API_URI || 'http://localhost:8080/'

export function getBlocks(): Promise<Array<ApiBlock>> {
    return apiCall(`blocks?limit=10`);
}

export function getBlock(hash: string): Promise<ApiBlock> {
    return apiCall(`block/${hash}`);
}

export function getTx(id: string): Promise<ApiTx> {
    return apiCall(`transaction/id/${id}`);
}

async function apiCall(path: string): Promise<any> {
    const response = await fetch(baseEndpoint + path).then(res => res.json());
    if (response.errorMessage != undefined) {
        throw response
    }
    return response
}

export interface ApiTx {
    transactionId: string
    transactionHash: string
    outputs: Array<{
        value: number
        address: string
    }>
    inputs: Array<{
        value: number
        address: string
    }>
    blocks: Array<ApiBlock>
}

export interface ApiBlock {
    blockHash: string
    blueScore: number
    timestamp: number
    parentBlockHashes: Array<string>
    transactionCount: number
    hashMerkleRoot: string
    acceptedIDMerkleRoot: string
    utxoCommitment: string
    version: number
    bits: number
    nonce: number
    difficulty: number
    transactionIds: Array<string>
}
