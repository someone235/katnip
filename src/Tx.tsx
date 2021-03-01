import React, {useEffect, useState} from 'react';
import {makeStyles} from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import Box from '@material-ui/core/Box';
import Typography from '@material-ui/core/Typography';
import Link from '@material-ui/core/Link';
import {useParams} from "react-router-dom";
import {ApiBlock, ApiTx, getTx} from "./lib/api";


const useStyles = makeStyles({
    table: {
        minWidth: 650,
    },
});

export default function TxPage() {
    const {id} = useParams<{ id: string }>()
    const [tx, setTx] = useState<ApiTx | undefined>()

    useEffect(() => {
        getTx(id).then(setTx)
    }, [id])

    if (tx == undefined) {
        return <div/>
    }

    const recipientAddresses = tx.outputs.map(output => output.address)
    const fromAddresses = tx.inputs.map(input => input.address)

    return (
        <Box>
            <Tx tx={tx}/>
            <RecipientAddresses addresses={recipientAddresses}/>
            {fromAddresses.length !== 0 ? <FromAddresses addresses={fromAddresses}/> : undefined}
            <IncludingBlocks blocks={tx.blocks}/>
        </Box>
    );
}

function IncludingBlocks({blocks}: { blocks: Array<ApiBlock> }) {
    const classes = useStyles();

    return (
        <Box my={4}>
            <Typography variant="h6" component="h1" gutterBottom>
                Including Blocks
            </Typography>
            <TableContainer component={Paper}>
                <Table className={classes.table} aria-label="Including blocks">
                    <TableHead>
                        <TableRow>
                            <TableCell>Hash</TableCell>
                            <TableCell>Blue Score</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {blocks.map(block =>
                            <TableRow key={block.blockHash}>
                                <TableCell><Link
                                    href={"#/block/" + block.blockHash}
                                    color="textSecondary">{block.blockHash}</Link></TableCell>
                                <TableCell>{block.blueScore}</TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </TableContainer>
        </Box>
    );
}

function Tx({tx}: { tx: ApiTx }) {
    const classes = useStyles();

    const unitsInKas = 1e8
    const totalInput = tx.inputs.reduce((sum, input) => sum + input.value, 0) / unitsInKas;
    const totalOutput = tx.outputs.reduce((sum, output) => sum + output.value, 0) / unitsInKas;

    const isCoinbase = tx.inputs.length === 0;
    const fee = isCoinbase ? 0 : totalInput - totalOutput

    return (
        <Box my={4}>
            <Typography variant="h6" component="h1" gutterBottom>
                Transaction {tx.transactionId}
            </Typography>
            <TableContainer component={Paper}>
                <Table className={classes.table} aria-label="Transaction">
                    <TableRow>
                        <TableCell variant="head">ID</TableCell>
                        <TableCell>{tx.transactionId}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Hash</TableCell>
                        <TableCell>{tx.transactionHash}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Total Input</TableCell>
                        <TableCell>
                            {totalInput} KAS
                        </TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Total Output</TableCell>
                        <TableCell>
                            {totalOutput} KAS
                        </TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Fees</TableCell>
                        <TableCell>{fee}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Number of Inputs</TableCell>
                        <TableCell>{tx.inputs.length}</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell variant="head">Number of Outputs</TableCell>
                        <TableCell>{tx.outputs.length}</TableCell>
                    </TableRow>
                </Table>
            </TableContainer>
        </Box>
    );
}

function RecipientAddresses({addresses}: { addresses: Array<string> }) {
    const classes = useStyles();

    return (
        <Box my={4}>
            <Typography variant="h6" component="h1" gutterBottom>
                Recipient Addresses
            </Typography>
            <TableContainer component={Paper}>
                <Table className={classes.table} aria-label="Recipient Addresses">
                    {addresses.map(address =>
                        <TableRow key={address}>
                            <TableCell><Typography
                                variant={"body2"}
                                color="textSecondary">{address}</Typography></TableCell>
                        </TableRow>
                    )}
                </Table>
            </TableContainer>
        </Box>
    );
}

function FromAddresses({addresses}: { addresses: Array<string> }) {
    const classes = useStyles();

    return (
        <Box my={4}>
            <Typography variant="h6" component="h1" gutterBottom>
                From Addresses
            </Typography>
            <TableContainer component={Paper}>
                <Table className={classes.table} aria-label="From Addresses">
                    {addresses.map(address =>
                        <TableRow key={address}>
                            <TableCell><Link
                                href="#"
                                color="textSecondary">{address}</Link></TableCell>
                        </TableRow>
                    )}
                </Table>
            </TableContainer>
        </Box>
    );
}