import React, {useEffect, useState} from 'react';
import {makeStyles, useTheme} from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import Box from '@material-ui/core/Box';
import Typography from '@material-ui/core/Typography';
import Link from '@material-ui/core/Link';
import {
    useParams
} from "react-router-dom";
import {ApiBlock, getBlock} from "./lib/api";


const useStyles = makeStyles({
    table: {
        minWidth: 650,
    },
});

export default function BlockPage() {
    const {hash} = useParams<{ hash: string }>()
    const [block, setBlock] = useState<ApiBlock | undefined>()

    useEffect(() => {
        getBlock(hash).then(setBlock)
    }, [hash])

    if (block == undefined) {
        return <div/>
    }

    return (
        <Box>
            <Block block={block}/>
            <Parents parents={block.parentBlockHashes}/>
            <Transactions transactionIds={block.transactionIds}/>
        </Box>
    );
}

function Parents({parents}: { parents: Array<string> }) {
    const classes = useStyles();

    return (
        <Box my={4}>
            <Typography variant="h6" component="h1" gutterBottom>
                Parents
            </Typography>
            <TableContainer component={Paper}>
                <Table className={classes.table} aria-label="Parents">
                    {parents.map(parent => <TableRow key={parent}>
                            <TableCell>
                                <Link
                                    color="textSecondary"
                                    href={`#/block/${parent}`}>{parent}</Link>
                            </TableCell>
                        </TableRow>
                    )}
                </Table>
            </TableContainer>
        </Box>
    );
}

function Transactions({transactionIds}: { transactionIds: Array<string> }) {
    return (
        <Box my={4}>
            <Typography variant="h6" component="h1" gutterBottom>
                Block Transactions
            </Typography>
            {transactionIds.length == 0 ? <PrunedBlockMessage/> : <TransactionsTable transactionIds={transactionIds}/>}
        </Box>
    );
}

function PrunedBlockMessage() {
    const theme = useTheme();
    return (
        <Box style={{backgroundColor: theme.palette.background.paper}} padding={1}>
            <Typography>
                Cannot show transactions for pruned or header only block
            </Typography>
        </Box>
    );
}

function TransactionsTable({transactionIds}: { transactionIds: Array<string> }) {
    const classes = useStyles();

    return (
        <TableContainer component={Paper}>
            <Table className={classes.table} aria-label="Transactions">
                {transactionIds.map(txId =>
                    <TableRow key={txId}>
                        <TableCell><Link
                            href={"#/tx/" + txId}
                            color="textSecondary">{txId}</Link></TableCell>
                    </TableRow>
                )}
            </Table>
        </TableContainer>
    );
}


function Block({block}: { block: ApiBlock }) {
    const classes = useStyles();

    const dateStr = (new Date(block.timestamp * 1000)).toLocaleString()

    return (
        <Box my={4}>
            <Typography variant="h6" component="h1" gutterBottom style={{overflowWrap: 'break-word'}}>
                Block {block.blockHash}
            </Typography>
            <TableContainer component={Paper}>
                <Table className={classes.table} aria-label="Block">
                    <TableRow>
                        <TableCell variant="head">Hash</TableCell>
                        <TableCell>{block.blockHash}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Timestamp</TableCell>
                        <TableCell>{dateStr}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Blue Score</TableCell>
                        <TableCell>{block.blueScore}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Number of Parents</TableCell>
                        <TableCell>{block.parentBlockHashes.length}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Number of Transactions</TableCell>
                        <TableCell>{block.transactionCount}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Difficulty</TableCell>
                        <TableCell>{block.difficulty}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Hash Merkle Root</TableCell>
                        <TableCell>{block.hashMerkleRoot}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Accepted ID Merkle Root</TableCell>
                        <TableCell>{block.acceptedIDMerkleRoot}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">UTXO Commitment</TableCell>
                        <TableCell>{block.utxoCommitment}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Version</TableCell>
                        <TableCell>{block.version}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Bits</TableCell>
                        <TableCell>0x{block.bits.toString(16)}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Nonce</TableCell>
                        <TableCell>0x{block.nonce.toString(16)}</TableCell>
                    </TableRow>
                </Table>
            </TableContainer>
        </Box>
    );
}