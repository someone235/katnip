import Box from "@material-ui/core/Box";
import TextField from "@material-ui/core/TextField";
import {InputAdornment} from "@material-ui/core";
import SearchIcon from "@material-ui/icons/Search";
import React, {SyntheticEvent, useState} from "react";

export default function Search() {
    const [search, setSearch] = useState("");

    async function handleSubmit(e: SyntheticEvent) {
        e.preventDefault();
        window.location.href = "#/search/" + search
    }

    return <Box my={4} style={{
        width: '80%',
        minWidth: '350px',
        backgroundColor: "grey"
    }}>
        <form noValidate autoComplete="off" onSubmit={handleSubmit}>
            <TextField onChange={e => setSearch(e.target.value)} style={{
                width: '100%'
            }}
                       InputLabelProps={{
                           style: {marginLeft: 10, color: 'white'}
                       }}
                       id="standard-basic" label="Search for block hash or transaction ID" InputProps={{
                endAdornment: (
                    <InputAdornment position="start">
                        <SearchIcon style={{cursor: 'pointer'}} onClick={handleSubmit}/>
                    </InputAdornment>
                ),
            }}/>
        </form>
    </Box>
}