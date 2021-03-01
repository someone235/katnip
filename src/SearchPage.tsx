import React, {useEffect, useState} from "react";
import {getBlock, getTx} from "./lib/api";
import {useParams} from "react-router-dom";
import Container from "@material-ui/core/Container";

export default function Search() {
    const {search} = useParams<{ search: string }>()
    const [err, setErr] = useState("");

    useEffect(() => {
        (async () => {
            try {
                await getBlock(search)
                window.location.href = "#/block/" + search
                return
            } catch (e: any) {
                if (e.errorCode == 404) {
                    try {
                        await getTx(search)
                        window.location.href = "#/tx/" + search
                        return
                    } catch (e: any) {
                        if (e.errorCode == 404) {
                            setErr(`Couldn't find a block or transaction with ID or hash of ${search}`)
                            return
                        }
                        setErr(`Error ${e.errorCode}: ${e.errorMessage}`)
                    }
                    return
                }
                setErr(`Error ${e.errorCode}: ${e.errorMessage}`)
            }
        })()
    }, [search])

    return <Container>
        {err}
    </Container>
}