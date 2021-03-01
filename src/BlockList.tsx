import React, {useState, useEffect} from 'react';
import {makeStyles} from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import Link from '@material-ui/core/Link';
import {getBlocks, ApiBlock} from "./lib/api";


const useStyles = makeStyles({
    table: {
        minWidth: 650,
    },
});

export default function BlockList() {
    const classes = useStyles();

    const [blocks, setBlocks] = useState<Array<ApiBlock>>([]);

    useEffect(() => {
        getBlocks().then(setBlocks)
        const timeoutID = setInterval(() => getBlocks().then(setBlocks), 1000);
        return () => clearTimeout(timeoutID)
    }, [null])

    return (
        <TableContainer component={Paper}>
            <Table className={classes.table} aria-label="Block list">
                <TableHead>
                    <TableRow>
                        <TableCell align={"center"}>Hash</TableCell>
                        <TableCell align={"center"}>Blue Score</TableCell>
                        <TableCell align={"center"}>Timestamp</TableCell>
                        <TableCell align={"center"}>Transactions</TableCell>
                        <TableCell  align={"center"} style={{maxWidth: "50px"}}>Number of parents</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {blocks.map((block) => {
                        const date = new Date(block.timestamp * 1000);
                        const dateStr = date.toLocaleString();
                        return <TableRow key={block.blockHash}>
                            <TableCell  align={"center"} component="th" scope="row" style={{width: 50}}>
                                < Link color="textSecondary" href={`#block/${block.blockHash}`}>{block.blockHash.substr(0, 32)}...</Link>
                            </TableCell>
                            <TableCell align={"center"}>{block.blueScore}</TableCell>
                            <TableCell align={"center"}>{dateStr}</TableCell>
                            <TableCell align={"center"}>{block.transactionCount}</TableCell>
                            <TableCell align={"center"}>{block.parentBlockHashes.length}</TableCell>
                        </TableRow>
                    })}
                </TableBody>
            </Table>
        </TableContainer>
    );
}
