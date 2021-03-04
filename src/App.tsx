import React from 'react';
import Container from '@material-ui/core/Container';
import Typography from '@material-ui/core/Typography';
import Box from '@material-ui/core/Box';
import Link from '@material-ui/core/Link';
import {createMuiTheme, CssBaseline} from '@material-ui/core';
import {ThemeProvider} from "@material-ui/styles";
import {
    HashRouter,
    Route,
} from "react-router-dom";

import Home from './Home';
import Block from './Block';
import Tx from './Tx';
import Search from './Search';
import SearchPage from './SearchPage';


const theme = createMuiTheme({
    palette: {
        type: "dark"
    }
});

export default function App() {
    return (
        <ThemeProvider theme={theme}>
            <CssBaseline/>
            <Container maxWidth="lg">
                <Box my={4}>
                    <Typography variant="h4" component="h1" gutterBottom>
                        <Link href={"/"}><img style={{
                            width: 50,
                            height: 50,
                        }} alt={""} src={"/img/Phoenician_kaph.svg"}/></Link> Katnip - Kaspa Block
                        Explorer
                    </Typography>
                    <Search/>
                    <HashRouter>
                        <Route exact path="/">
                            <Home/>
                        </Route>
                        <Route path="/block/:hash">
                            <Block/>
                        </Route>
                        <Route path="/tx/:id">
                            <Tx/>
                        </Route>
                        <Route path="/search/:search">
                            <SearchPage/>
                        </Route>
                    </HashRouter>
                </Box>
            </Container>
        </ThemeProvider>
    );
}
